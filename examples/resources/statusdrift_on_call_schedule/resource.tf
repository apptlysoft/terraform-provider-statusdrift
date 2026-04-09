resource "statusdrift_on_call_schedule" "primary" {
  name              = "Primary On-Call"
  timezone          = "America/New_York"
  frequency         = "weekly"
  rotation_interval = 1
  starts_at         = "2024-01-01T00:00:00Z"
  enabled           = true
  participant_ids   = [1, 2, 3]
}
