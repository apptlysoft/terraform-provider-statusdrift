resource "statusdrift_notification_channel" "slack_alerts" {
  name    = "Slack - #alerts"
  type    = "slack"
  enabled = true

  config = {
    webhook_url = "https://hooks.slack.com/services/T00/B00/xxxx"
  }
}

resource "statusdrift_notification_channel" "email_oncall" {
  name    = "On-Call Email"
  type    = "email"
  enabled = true

  config = {
    email = "oncall@example.com"
  }
}
