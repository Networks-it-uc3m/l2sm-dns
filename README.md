# Go Project Template

This is a Go project template that provides a standard layout for CLI applications and/or services. It includes:

- **cmd/** for CLI entrypoints
- **api/** for API definitions (e.g., .proto files)
- **pkg/** for exportable/reusable packages
- **internal/** for private/internal packages
- **config/** for KubeBuilder or other configuration files
- **test/** for end-to-end or integration tests
- A **Makefile** for common build tasks
- GitHub Actions workflow in `.github/workflows/ci.yaml` for CI

## Getting Started

1. **Clone the Template**  
   ```bash
   git clone https://github.com/<your-org>/go-project-template.git
   cd go-project-template
   ```

2. **Update go.mod**  
   Change `module github.com/<your-org>/go-project-template` to your own module path:
   ```bash
   go mod edit -module github.com/<your-org>/your-new-service
   go mod tidy
   ```

3. **Build**  
   ```bash
   make build
   ```

4. **Test**  
   ```bash
   make test
   ```

5. **Run the CLI**  
   ```bash
   ./bin/yourapp
   ```

