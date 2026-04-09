resource "statusdrift_incident" "api_outage" {
  monitor_id = statusdrift_monitor.api_health.id
  message    = "API is returning 500 errors intermittently"
  priority   = "high"
}
