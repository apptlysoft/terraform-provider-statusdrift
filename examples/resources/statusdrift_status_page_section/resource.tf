resource "statusdrift_status_page_section" "api" {
  status_page_id = statusdrift_status_page.public.id
  name           = "API Services"
  description    = "Core API endpoints"
  position       = 0
  expanded       = true
  enabled        = true
}
