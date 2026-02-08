# TaskMate Provider Example
# 
# Before running:
# 1. Start TaskMate API: 
#    git clone https://github.com/tyagian/taskmate
#    cd taskmate && go run main.go
# 2. Generate token: curl -X POST http://localhost:8080/api/v1/auth/token
# 3. Set token: export TF_VAR_taskmate_token="your-token"
# 4. Run: terraform plan && terraform apply

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
  token = var.taskmate_token  # Set via: export TF_VAR_taskmate_token="your-token"
}

variable "taskmate_token" {
  type        = string
  description = "TaskMate API token"
  sensitive   = true
}

# Example 1: Create a task
resource "taskmate_task" "deployment" {
  title       = "Deploy to Production"
  description = "Deploy v2.0 release"
  due_date    = "2026-02-31"
  priority    = "high"
  status      = "pending"
}

# Example 2: Create another task
resource "taskmate_task" "testing" {
  title       = "Run Integration Tests"
  description = "Test all endpoints"
  due_date    = "2026-02-22"
  priority    = "medium"
  status      = "pending"
}

# Example 3: Query a specific task (data source)
data "taskmate_task" "check_deployment" {
  id = taskmate_task.deployment.id
}

# Example 4: List all tasks (data source)
data "taskmate_tasks" "all" {}

# Outputs
output "deployment_task_id" {
  description = "ID of the deployment task"
  value       = taskmate_task.deployment.id
}

output "deployment_task_title" {
  description = "Title of the deployment task"
  value       = taskmate_task.deployment.title
}

output "total_tasks" {
  description = "Total number of tasks"
  value       = length(data.taskmate_tasks.all.tasks)
}

output "high_priority_tasks" {
  description = "List of high priority tasks"
  value = [
    for task in data.taskmate_tasks.all.tasks : task.title
    if task.priority == "high"
  ]
}
