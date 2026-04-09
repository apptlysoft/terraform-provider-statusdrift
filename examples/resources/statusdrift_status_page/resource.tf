resource "statusdrift_status_page" "public" {
  name        = "Service Status"
  type        = "all"
  enabled     = true
  description = "Current status of all services"

  show_announcements = true
  show_incidents     = true
  show_history       = true
  history_days       = 90
  show_uptime        = true
  show_response_time = true
  show_overall_status = true

  color_theme  = "light"
  timezone     = "America/New_York"
}
