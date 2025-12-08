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

// note that we want to pass in descriptions of the nested fields to the LLM so we can do that with the jsonschema annotation
type SomeJsonObject struct {
	Foo    string `json:"foo" jsonschema:"Description of Foo field"`
	Bar    int `json:"foo" jsonschema:"Description of Bar field"`
	FooBar struct {
		Baz string `json:"baz" jsonschema:"Description of Baz field"`
	} `json:"foo_bar" jsonschema:"Description of FooBar field"`
}

// generate schema for our object
aJsonObjSchema, err := jsonschema.For[SomeJsonObject](nil)
if err != nil {
	// do something better than this in prod
	panic(err)
}
// we want a description on the schema for the LLM to use and we'd like to re-use it for Cobra too
flagDescription := "This flag is used to supply a JSON string representation of an instance of a SomeJsonObject"
aJsoObjSchema.Description = flagDescription
bytes, err := aJsonObjSchema.MarshalJSON()
if err != nil {
    // do something better than this in prod
    panic(err)
}
// now create flag that has a json schema that represents a json object
cmd.Flags().String("a_json_obj", "", flagDescription)
jsonobj := cmd.Flags().Lookup("a_json_obj")
jsonobj.Annotations = make(map[string][]string)
jsonobj.Annotations["jsonschema"] = []string{string(bytes)}

// within the Cobra environment, fetch the flag value then deserialize into your struct
var myFlagObj SomeJsonObject
value, err := cmd.Flags().GetString("a_json_obj")
if err != nil {
	panic(err)
}
err = json.Unmarshal([]byte(value), &myFlagObj)
if err != nil {
	panic(err)
}
// now do something with your object

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
