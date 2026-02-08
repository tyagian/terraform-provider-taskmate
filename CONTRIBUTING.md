# Contributing to Terraform Provider TaskMate

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to the TaskMate Terraform provider.

## Development Environment Setup

### Prerequisites

- [Go](https://golang.org/doc/install) >= 1.21
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [golangci-lint](https://golangci-lint.run/usage/install/) (for linting)
- [TaskMate API](https://github.com/tyagian/taskmate) running locally

### Getting Started

1. Fork and clone the repository:
```bash
git clone https://github.com/tyagian/terraform-provider-taskmate.git
cd terraform-provider-taskmate
```

2. Install dependencies:
```bash
go mod download
```

3. Build the provider:
```bash
make build
```

## Development Workflow

### Building

```bash
# Build the provider
make build

# Install locally for testing
make install
```

### Testing

```bash
# Run unit tests
make test

# Run acceptance tests (requires TaskMate API running)
make testacc

# Run with coverage
go test ./... -cover
```

### Linting

```bash
# Run linters
make lint

# Format code
make fmt
```

### Documentation

```bash
# Generate documentation
make docs
```

This uses `tfplugindocs` to generate documentation from:
- Resource/data source schemas
- Example files in `examples/`
- Templates in `templates/` (if any)

## Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Ensure all linters pass
- Add comments for exported functions and types
- Keep functions focused and testable

## Testing Guidelines

### Unit Tests

- Test files should be named `*_test.go`
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for good coverage of edge cases

Example:
```go
func TestTaskResourceCreate(t *testing.T) {
    tests := []struct {
        name    string
        input   TaskModel
        wantErr bool
    }{
        {
            name: "valid task",
            input: TaskModel{
                Title: types.StringValue("Test"),
            },
            wantErr: false,
        },
        // Add more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Acceptance Tests

Acceptance tests require a running TaskMate API instance:

```go
func TestAccTaskResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Test steps
        },
    })
}
```

## Pull Request Process

1. Create a feature branch:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes:
   - Write code
   - Add tests
   - Update documentation
   - Run linters and tests

3. Commit with clear messages:
```bash
git commit -m "Add feature: description of what you added"
```

4. Push and create a pull request:
```bash
git push origin feature/your-feature-name
```

5. PR checklist:
   - [ ] Tests pass (`make test`)
   - [ ] Linters pass (`make lint`)
   - [ ] Documentation updated
   - [ ] Examples added/updated if needed
   - [ ] Clear description of changes

## Adding New Resources

When adding a new resource:

1. Create the resource file in `internal/provider/`:
```go
// internal/provider/new_resource.go
package provider

type NewResourceModel struct {
    // Define schema
}

func (r *NewResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    // Define schema
}

// Implement CRUD methods
```

2. Register in provider:
```go
// internal/provider/provider.go
func (p *TaskMateProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewTaskResource,
        NewNewResource, // Add here
    }
}
```

3. Add examples in `examples/resources/taskmate_new_resource/`

4. Add tests in `internal/provider/new_resource_test.go`

5. Generate docs: `make docs`

## Adding New Data Sources

Similar to resources, but implement `datasource.DataSource` interface:

1. Create data source file
2. Register in provider's `DataSources()` method
3. Add examples
4. Add tests
5. Generate docs

## API Client Changes

When modifying the API client (`internal/provider/client.go`):

1. Keep methods focused and single-purpose
2. Handle errors appropriately
3. Add proper logging
4. Update tests
5. Document any breaking changes

## Documentation

- Keep README.md up to date
- Add examples for new features
- Use clear, concise language
- Include code snippets
- Document any breaking changes in CHANGELOG.md

## Release Process

Releases are automated via GitHub Actions:

1. Update CHANGELOG.md
2. Create and push a version tag:
```bash
git tag v0.2.0
git push origin v0.2.0
```

3. GitHub Actions will:
   - Run tests
   - Build binaries for all platforms
   - Create GitHub release
   - Sign artifacts with GPG

## Getting Help

- Open an issue for bugs or feature requests
- Check existing issues and PRs
- Ask questions in discussions

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## License

By contributing, you agree that your contributions will be licensed under the Mozilla Public License 2.0.
