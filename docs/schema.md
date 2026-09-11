# Schema Generation

Ophis automatically generates JSON schemas for MCP tools from Cobra commands.

## Tool Properties

- **Name**: Command path with underscores (`kubectl_get_pods`)
- **Description**: From command's Long, Short, and Example fields
- **Input Schema**: Generated from flags and arguments
- **Output Schema**: Standard format (stdout, stderr, exitCode)

## Input Schema

### Flags

Each flag becomes a property with:

**Type mapping:**

- `bool` → `boolean`
- `int`, `uint` → `integer`
- `float` → `number`
- `string` → `string`
- any flag type → `<JSON schema>` if the flag has an annotation called `jsonschema` with a value that is the JSON string representation of the schema
- `stringSlice`, `intSlice` → `array`
- `duration` → `string` with pattern validation
- `ip`, `ipNet` → `string`; pflag validates their values

Flags marked as required (via `cmd.MarkFlagRequired()`) are included in the schema's `required` array. Default values are included in the schema, except for empty strings (`""`) and empty arrays (`[]`). A valid default supplied by a `jsonschema` annotation takes precedence over the pflag default.

**Example:**

```json
{
  "flags": {
    "type": "object",
    "properties": {
      "namespace": {
        "type": "string",
        "description": "Kubernetes namespace",
        "default": "default"
      },
      "replicas": {
        "type": "integer",
        "default": 3
      },
      "labels": {
        "type": "array",
        "items": { "type": "string" }
      }
    },
    "required": ["namespace"]
  }
}
```

Example showing JSON schema:

```golang

type SomeJsonObject struct {
	Foo    string
	Bar    int
	FooBar struct {
		Baz string
	}
}

// generate schema for our object
aJsonObjSchema, err := jsonschema.For[SomeJsonObject](nil)
if err != nil {
	// do something better than this in prod
	panic(err)
}
bytes, err := aJsonObjSchema.MarshalJSON()
if err != nil {
    // do something better than this in prod
    panic(err)
}
// now create flag that has a json schema that represents a json object
cmd.Flags().String("a_json_obj", "", "Some JSON Object")
jsonobj := cmd.Flags().Lookup("a_json_obj")
jsonobj.Annotations = make(map[string][]string)
jsonobj.Annotations[ophis.FlagAnnotationJSONSchema] = []string{string(bytes)}

```

### Arguments

Positional arguments are a string array. When `Config.InferArgConstraints` is enabled, required, optional, and variadic argument counts are derived from recognizable Cobra `Use` patterns and emitted as `required`, `minItems`, and `maxItems` constraints when corroborated by the command's `Args` validator. Ambiguous, undocumented, or unenforced bounds remain unconstrained.

Constraint inference invokes each command's `Args` validator with synthetic placeholder arguments during tool registration. Enable it only for validators that are pure and depend only on the number of arguments, not their values or flag state. Output written through the probed command is discarded, and panics leave argument bounds unconstrained; direct process I/O and exits cannot be contained.

```json
{
  "args": {
    "type": "array",
    "description": "Positional command line arguments\nUsage pattern: [NAME]",
    "items": { "type": "string" },
    "maxItems": 1
  }
}
```

## Output Schema

```json
{
  "type": "object",
  "properties": {
    "stdout": { "type": "string" },
    "stderr": { "type": "string" },
    "exitCode": { "type": "integer" }
  }
}
```

## Export Schemas

```bash
./my-cli mcp tools  # Creates mcp-tools.json
```
