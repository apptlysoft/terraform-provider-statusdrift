resource "statusdrift_sla_policy" "production" {
  name          = "Production SLA"
  uptime_target = 99.9
  window_type   = "calendar_month"
  scope         = "org"

  response_time_sla_enabled  = true
  response_time_threshold_ms = 500

  monitor_ids = [statusdrift_monitor.https.id]
  tag_ids     = [statusdrift_tag.critical.id]
}
