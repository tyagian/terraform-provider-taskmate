resource "taskmate_task" "example" {
  title       = "Deploy to Production"
  description = "Deploy v2.0 release"
  due_date    = "2024-12-31"
  priority    = "high"
  status      = "pending"
}
