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
- `string` → `<JSON schema>` if the flag has an annotation called `jsonschema` with a value that is the JSON string representation of the schema
- `stringSlice`, `intSlice` → `array`
- `duration`, `ip`, `ipNet` → `string` with pattern validation

Flags marked as required (via `cmd.MarkFlagRequired()`) are included in the schema's `required` array. Default values are included in the schema, except for empty strings (`""`) and empty arrays (`[]`).

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
jsonobj.Annotations["jsonschema"] = []string{string(bytes)}

```

### Arguments

Positional arguments are a string array:

```json
{
  "args": {
    "type": "array",
    "description": "Positional arguments\nUsage: [NAME] [flags]",
    "items": { "type": "string" }
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
