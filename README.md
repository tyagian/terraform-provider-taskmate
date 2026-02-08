# Terraform Provider for TaskMate

[![Tests](https://github.com/tyagian/terraform-provider-taskmate/actions/workflows/test.yml/badge.svg)](https://github.com/tyagian/terraform-provider-taskmate/actions/workflows/test.yml)
[![Release](https://github.com/tyagian/terraform-provider-taskmate/actions/workflows/release.yml/badge.svg)](https://github.com/tyagian/terraform-provider-taskmate/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/tyagian/terraform-provider-taskmate)](https://goreportcard.com/report/github.com/tyagian/terraform-provider-taskmate)
[![License](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](LICENSE)

A Terraform provider for managing tasks through the [TaskMate API](https://github.com/tyagian/taskmate). Built with the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) for educational purposes to develop terraform providers

## Features

- ✅ Full CRUD operations for tasks
- ✅ Data sources for querying single or multiple tasks
- ✅ Import support for existing tasks
- ✅ Token-based authentication
- ✅ Comprehensive examples and documentation
- ✅ Production-ready with CI/CD workflows

## Quick Start

### 1. Install TaskMate API

```bash
git clone https://github.com/tyagian/taskmate
cd taskmate
go run main.go
```

The API will start on `http://localhost:8080`.

### 2. Generate API Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/token
```

Save the returned token - you'll need it for authentication.

### 3. Configure the Provider

```hcl
terraform {
  required_providers {
    taskmate = {
      source = "tyagian/taskmate"
      version = "~> 1.0"
    }
  }
}

provider "taskmate" {
  host  = "http://localhost:8080"
  token = var.taskmate_token  # Or use TASKMATE_TOKEN env var
}
```

### 4. Create Your First Task

```hcl
resource "taskmate_task" "example" {
  title       = "Deploy Application"
  description = "Deploy v2.0 to production"
  priority    = "high"
  due_date    = "2026-02-31"
}

output "task_id" {
  value = taskmate_task.example.id
}
```

Run:
```bash
terraform init
terraform apply
```

## Installation

### From Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    taskmate = {
      source  = "tyagian/taskmate"
      version = "~> 1.0"
    }
  }
}
```

### Local Development Installation

For local development and testing:

```bash
# Clone the repository
git clone https://github.com/tyagian/terraform-provider-taskmate
cd terraform-provider-taskmate

# Build and install
make install
```

Or use dev overrides in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "tyagian/taskmate" = "/absolute/path/to/terraform-provider-taskmate/bin"
  }
  direct {}
}
```

## Authentication

TaskMate uses token-based authentication for write operations (POST, PUT, DELETE). Read operations (GET) don't require authentication.

### Generate a Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/token
```

**Response:**
```json
{
  "token": "a1b2c3d4e5f6...",
  "message": "Token generated successfully. Save this token securely."
}
```

### Configure Authentication

**Option 1: Environment Variable (Recommended)**
```bash
export TASKMATE_TOKEN="your-generated-token"
```

**Option 2: Provider Configuration**
```hcl
provider "taskmate" {
  host  = "http://localhost:8080"
  token = "your-generated-token"
}
```

**Option 3: Variable**
```hcl
variable "taskmate_token" {
  type      = string
  sensitive = true
}

provider "taskmate" {
  host  = "http://localhost:8080"
  token = var.taskmate_token
}
```

## Usage Examples

### Managing Tasks

```hcl
# Create a task
resource "taskmate_task" "deployment" {
  title       = "Deploy to Production"
  description = "Deploy v2.0 release"
  priority    = "high"
  due_date    = "2024-12-31"
}

# Query a specific task
data "taskmate_task" "existing" {
  id = "1"
}

# List all tasks
data "taskmate_tasks" "all" {}

# Filter tasks
output "high_priority_tasks" {
  value = [
    for task in data.taskmate_tasks.all.tasks : task
    if task.priority == "high"
  ]
}
```

### Import Existing Tasks

```bash
# Discover task IDs
curl http://localhost:8080/api/v1/tasks | jq '.[] | .id'

# Import a task
terraform import taskmate_task.imported 1
```

See [examples/import/](examples/import/) for detailed import workflows.

## Documentation

- [Provider Configuration](docs/index.md)
- [Resources](docs/resources/)
  - [taskmate_task](docs/resources/task.md)
- [Data Sources](docs/data-sources/)
  - [taskmate_task](docs/data-sources/task.md)
  - [taskmate_tasks](docs/data-sources/tasks.md)
- [Examples](examples/)

## Development

### Prerequisites

- [Go](https://golang.org/doc/install) >= 1.21
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [golangci-lint](https://golangci-lint.run/usage/install/)
- [TaskMate API](https://github.com/tyagian/taskmate) running locally

### Building

```bash
# Download dependencies
go mod download

# Build the provider
make build

# Install locally
make install
```

### Testing

```bash
# Run unit tests
make test

# Run acceptance tests (requires TaskMate API)
make testacc

# Run linters
make lint

# Format code
make fmt
```

### Generating Documentation

```bash
# Generate provider documentation
make docs
```

This uses [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) to generate documentation from:
- Resource and data source schemas
- Example files in `examples/`
- Description fields in the code

### Project Structure

```
terraform-provider-taskmate/
├── .github/
│   └── workflows/          # CI/CD workflows
│       ├── test.yml        # Run tests on PRs
│       └── release.yml     # Release automation
├── docs/                   # Generated documentation
├── examples/               # Example configurations
│   ├── provider/          # Provider configuration
│   ├── resources/         # Resource examples
│   ├── data-sources/      # Data source examples
│   └── import/            # Import examples
├── internal/
│   └── provider/          # Provider implementation
│       ├── provider.go    # Provider configuration
│       ├── client.go      # API client
│       ├── task_resource.go
│       ├── task_data_source.go
│       └── tasks_data_source.go
├── tools/                 # Build tools
│   └── tools.go          # Tool dependencies
├── .golangci.yml         # Linter configuration
├── .goreleaser.yml       # Release configuration
├── Makefile              # Build automation
├── main.go               # Provider entry point
└── README.md             # This file
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Quick Contribution Guide

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `make test`
5. Run linters: `make lint`
6. Commit: `git commit -m 'Add amazing feature'`
7. Push: `git push origin feature/amazing-feature`
8. Open a Pull Request

## Release Process

Releases are automated via GitHub Actions:

1. Update CHANGELOG.md with release notes
2. Create and push a version tag:
   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```
3. GitHub Actions will automatically:
   - Run tests
   - Build binaries for all platforms
   - Create a GitHub release
   - Sign artifacts with GPG
   - Upload to Terraform Registry (if configured)

## Troubleshooting

### "Token required" error

Generate a new token:
```bash
curl -X POST http://localhost:8080/api/v1/auth/token
```

### "Connection refused"

Ensure TaskMate API is running:
```bash
curl http://localhost:8080/health
```

### Provider not found

Check your Terraform configuration and ensure the provider is properly installed:
```bash
terraform init
```

For local development, verify your `~/.terraformrc` dev override path is correct.

### Debug Mode

Enable debug logging:
```bash
export TF_LOG=DEBUG
terraform apply
```

## API Compatibility

This provider is compatible with TaskMate API v1.0+:

- **Authentication:** Token-based (X-API-Token header)
- **Read operations:** No authentication required
- **Write operations:** Token required (POST, PUT, DELETE)

## License

Mozilla Public License 2.0 - see [LICENSE](LICENSE) for details.

## Resources

- [TaskMate API](https://github.com/tyagian/taskmate)
- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Terraform Provider Development](https://developer.hashicorp.com/terraform/plugin)
- [Terraform Registry](https://registry.terraform.io/)

## Acknowledgments

Built with the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) and inspired by HashiCorp's [provider scaffolding template](https://github.com/hashicorp/terraform-provider-scaffolding-framework).

---

**Note:** This provider is designed for educational purposes to demonstrate Terraform provider development. It's production-ready but intended as a learning resource for building custom Terraform providers.
