resource "statusdrift_status_page_announcement" "maintenance_notice" {
  status_page_id = statusdrift_status_page.public.id
  title          = "Scheduled Maintenance"
  content        = "We will be performing scheduled maintenance on our API servers."
  type           = "maintenance"
  enabled        = true
  starts_at      = "2025-03-20T02:00:00Z"
  ends_at        = "2025-03-20T06:00:00Z"
}
