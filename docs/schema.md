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
