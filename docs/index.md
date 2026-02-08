---
page_title: "TaskMate Provider"
subcategory: ""
description: |-
  The TaskMate provider is used to interact with TaskMate API resources.
---

# TaskMate Provider

The TaskMate provider allows Terraform to manage tasks through the TaskMate API. This provider is built with the Terraform Plugin Framework and supports full CRUD operations, data sources, and import functionality.

## Example Usage

```terraform
terraform {
  required_providers {
    taskmate = {
      source  = "tyagian/taskmate"
      version = "~> 1.0"
    }
  }
}

provider "taskmate" {
  host  = "http://localhost:8080"
  token = var.taskmate_token
}

variable "taskmate_token" {
  type        = string
  description = "TaskMate API token"
  sensitive   = true
}
```

## Authentication

TaskMate uses token-based authentication for write operations (POST, PUT, DELETE). Read operations (GET) don't require authentication.

### Generate a Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/token
```

The response will contain a token that you should save securely:

```json
{
  "token": "a1b2c3d4e5f6...",
  "message": "Token generated successfully. Save this token securely."
}
```

### Configure Authentication

You can provide the token in three ways:

1. **Environment Variable (Recommended)**
   ```bash
   export TASKMATE_TOKEN="your-generated-token"
   ```

2. **Provider Configuration**
   ```hcl
   provider "taskmate" {
     token = "your-generated-token"
   }
   ```

3. **Terraform Variable**
   ```hcl
   variable "taskmate_token" {
     type      = string
     sensitive = true
   }

   provider "taskmate" {
     token = var.taskmate_token
   }
   ```

## Schema

### Required

- `host` (String) The TaskMate API host URL. Can also be set via the `TASKMATE_HOST` environment variable.

### Optional

- `token` (String, Sensitive) The API token for authentication. Can also be set via the `TASKMATE_TOKEN` environment variable. Required for write operations (create, update, delete).
