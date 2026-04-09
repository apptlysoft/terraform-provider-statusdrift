resource "statusdrift_on_call_policy" "default" {
  name       = "Default Escalation"
  enabled    = true
  is_default = true
  scope_type = "org_default"

  step {
    target_type   = "schedule"
    schedule_id   = statusdrift_on_call_schedule.primary.id
    delay_minutes = 0
  }

  step {
    target_type            = "member"
    organization_member_id = 5
    delay_minutes          = 5
  }
}
