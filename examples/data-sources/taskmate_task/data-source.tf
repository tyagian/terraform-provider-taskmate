data "taskmate_task" "example" {
  id = "1"
}

output "task_title" {
  value = data.taskmate_task.example.title
}
