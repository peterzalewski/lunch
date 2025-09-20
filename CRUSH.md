# CRUSH.md - Lunch CLI Tool

This is a toy project for learning idiomatic Go. New York State's Department of Education publishes monthly menus for public schools. I want `lunch`, this project, to download and parse those menus and then make it explorable via TUI but also potentially text subscribers the menu each morning to help them make a decision on whether to pack lunch or not.

## Build/Test/Lint Commands
- **Build**: `go build -o lunch .` or `go build .`
- **Run**: `go run main.go` or `./lunch`
- **Test**: `go test ./...` (no tests currently exist)
- **Single test**: `go test -run TestFunctionName ./...`
- **Vet**: `go vet ./...`
- **Format**: `go fmt ./...`
- **Mod tidy**: `go mod tidy`

## Code Style Guidelines

### Imports
- Standard library imports first, then third-party packages
- Use `github.com/adrg/xdg` for cross-platform cache directories

### Types & Naming
- Use PascalCase for exported types (LunchConfig, LunchOption)
- Use camelCase for unexported fields (schoolYear, basePath)
- Method receivers use abbreviated type names (lc for LunchConfig, lo for LunchOption)

### Error Handling
- Return errors as second return value: `func() (string, error)`
- Check errors immediately after function calls
- Use `defer` for cleanup (resp.Body.Close(), writer.Flush())

### Formatting
- Use `go fmt` for consistent formatting
- TODO comments for future improvements
- String building with `strings.Builder` for efficiency
- Use `fmt.Sprintf` for string formatting

### Project Structure
- Single main.go file for this simple CLI tool
- Types defined before functions
- Main function at bottom
