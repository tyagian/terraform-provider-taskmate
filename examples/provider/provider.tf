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
