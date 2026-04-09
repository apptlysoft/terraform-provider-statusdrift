# Look up a member by email for use in on-call schedules
data "statusdrift_collaborator" "alice" {
  email = "alice@example.com"
}

resource "statusdrift_on_call_schedule" "primary" {
  name              = "Primary On-Call"
  timezone          = "UTC"
  frequency         = "weekly"
  rotation_interval = 1
  starts_at         = "2025-01-01T00:00:00Z"
  participant_ids   = [data.statusdrift_collaborator.alice.id]
}
