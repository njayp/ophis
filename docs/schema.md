# Schema Generation

Ophis automatically generates JSON schemas for MCP tools from your Cobra commands.

## Command Discovery

Ophis recursively walks your Cobra command tree at runtime. For each command that matches a selector, it creates an MCP tool with:

- **Name**: Command path with spaces replaced by underscores (e.g., `kubectl_get_pods`)
- **Description**: Derived from command's Long, Short, and Example fields
- **Input Schema**: Generated from flags and arguments
- **Output Schema**: Standard output format

## Input Schema

### Flags

Each flag becomes a property in the `flags` object with:

- **Type**: Mapped from Cobra flag types
  - `bool` → `"boolean"`
  - `int`, `int32`, `int64`, `uint` → `"integer"`
  - `float32`, `float64` → `"number"`
  - `string` → `"string"`
  - `stringSlice`, `intSlice`, etc. → `"array"`
  - `duration` → `"string"` with pattern validation
  - `ip`, `ipNet` → `"string"` with pattern validation

- **Description**: From flag's Usage field
- **Default**: Set if flag has non-zero default value
- **Required**: Set if flag is marked required via `cmd.MarkFlagRequired()`

Example flag schema:

```json
{
  "namespace": {
    "type": "string",
    "description": "Kubernetes namespace",
    "default": "default"
  },
  "replicas": {
    "type": "integer",
    "description": "Number of replicas",
    "default": 3
  },
  "verbose": {
    "type": "boolean",
    "description": "Enable verbose output",
    "default": false
  },
  "labels": {
    "type": "array",
    "description": "Resource labels",
    "items": {
      "type": "string"
    }
  }
}
```

### Arguments

Positional arguments are represented as an array of strings:

```json
{
  "args": {
    "type": "array",
    "description": "Positional command line arguments\nUsage pattern: [NAME] [flags]",
    "items": {
      "type": "string"
    }
  }
}
```

The usage pattern is extracted from the command's `Use` field.

### Complete Input Example

```json
{
  "type": "object",
  "properties": {
    "flags": {
      "type": "object",
      "properties": {
        "namespace": { "type": "string", "default": "default" },
        "output": { "type": "string", "description": "Output format (json|yaml)" }
      },
      "required": ["namespace"]
    },
    "args": {
      "type": "array",
      "items": { "type": "string" }
    }
  }
}
```

## Output Schema

All tools return a consistent output structure:

```json
{
  "type": "object",
  "properties": {
    "stdout": {
      "type": "string",
      "description": "Standard output"
    },
    "stderr": {
      "type": "string",
      "description": "Standard error"
    },
    "exitCode": {
      "type": "integer",
      "description": "Exit code"
    }
  }
}
```

## Array Default Values

Array flags with default values are parsed from Cobra's bracket notation:

- `[hello,world]` → `["hello", "world"]`
- `[1,2,3]` → `[1, 2, 3]`
- `[]` → No default (empty arrays excluded)

Invalid array elements are logged as warnings and skipped.

## Special Flag Types

### Duration

```json
{
  "type": "string",
  "description": "Request timeout (format: Go duration string, e.g., '10s', '2h45m')",
  "pattern": "^-?([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
}
```

### IP Address

```json
{
  "type": "string",
  "description": "Server IP address (format: IPv4 or IPv6 address)",
  "pattern": "^((25[0-5]|(2[0-4]|1\\d|[1-9]|)\\d)\\.){3}(25[0-5]|(2[0-4]|1\\d|[1-9]|)\\d)$|^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4})$"
}
```

### CIDR Notation

```json
{
  "type": "string",
  "description": "Network CIDR (format: CIDR notation, e.g., '192.168.1.0/24')",
  "pattern": "^((25[0-5]|(2[0-4]|1\\d|[1-9]|)\\d)\\.){3}(25[0-5]|(2[0-4]|1\\d|[1-9]|)\\d)/([0-9]|[1-2][0-9]|3[0-2])$"
}
```

### Base64

```json
{
  "type": "string",
  "description": "Encoded data (format: base64 encoded string)",
  "pattern": "^[A-Za-z0-9+/]*={0,2}$"
}
```

## Schema Export

Export tool schemas for inspection:

```bash
./my-cli mcp tools
```

This creates `mcp-tools.json` with all tool definitions.
