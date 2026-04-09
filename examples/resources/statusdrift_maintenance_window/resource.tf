resource "statusdrift_maintenance_window" "weekly_deploy" {
  name        = "Weekly Deploy Window"
  repeat      = "weekly"
  enabled     = true
  target_type = "all"
  start_time  = "02:00"
  end_time    = "04:00"
  timezone    = "America/New_York"
  weekly_days = ["Tuesday", "Thursday"]
}

resource "statusdrift_maintenance_window" "one_time" {
  name            = "Database Migration"
  repeat          = "none"
  enabled         = true
  target_type     = "group"
  target_id       = statusdrift_monitor_group.production.id
  start_date_time = "2025-04-01T02:00:00Z"
  end_date_time   = "2025-04-01T06:00:00Z"
}
