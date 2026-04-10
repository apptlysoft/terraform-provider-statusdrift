resource "statusdrift_monitor" "api_health" {
  name          = "API Health Check"
  type          = "HTTPS"
  url           = "https://api.example.com/health"
  interval      = 5
  enabled       = true
  location_type = "region"
  region        = "us-east-1"

  http_method          = "GET"
  expected_status_code = "200"
  max_redirects        = 3
  rw_timeout           = 10
  connection_timeout   = 10

  monitor_group_id = statusdrift_monitor_group.production.id
  tag_ids          = [statusdrift_tag.critical.id]

  locations_down_to_alert = 1
  checks_down_to_alert    = 2
  warning_threshold_ms    = 5000

  alert {
    notification_channel_ids   = [statusdrift_notification_channel.slack_alerts.id]
    conditions                 = ["down", "warning"]
    delay_minutes              = 2
    enabled                    = true
    recurring_interval_minutes = 15
    max_recurrences            = 5
    notify_on_call             = true
    on_call_policy_id          = statusdrift_on_call_policy.default.id
  }

  assertion {
    type           = "json_path"
    expression     = "$.status"
    comparison     = "equals"
    expected_value = "healthy"
  }

  assertion {
    type       = "regex"
    expression = "\"status\"\\s*:\\s*\"healthy\""
    comparison = "matches"
  }
}

resource "statusdrift_monitor" "dns_check" {
  name                = "DNS Monitor"
  type                = "DNS"
  interval            = 60
  enabled             = true
  location_type       = "region"
  region              = "us-east-1"
  dns_hostname        = "example.com"
  dns_monitoring_mode = "discovery"
}
