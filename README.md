# Environment Configuration Loader

A Go package for loading environment variables from `.env` files and system environment into structs with support for various data types and custom parsing. A simple, zero-dependencies library to parse environment variables into structs.

## Features

- Load environment variables from `.env` files
- Support for system environment variables
- Type conversion for common Go types
- Support for nested structs
- Custom parsers for complex types
- Required and optional fields
- Default values
- Environment variable expansion
- Multi-line values support

## Installation

```bash
go get github.com/ealoshinsky/environment
```

## Supported Types

- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `bool`
- `time.Duration`
- `[]string` (comma-separated)
- `map[string]string` (JSON format)
- Custom types implementing `CustomParser` interface

## Tags

- `env` - Environment variable name
- `default` - Default value if environment variable is not set
- `required` - Set to "true" if the variable is required

## Error Handling

The package returns descriptive errors for various scenarios:

- Missing required variables
- Invalid variable names
- Type conversion errors
- File reading errors

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
