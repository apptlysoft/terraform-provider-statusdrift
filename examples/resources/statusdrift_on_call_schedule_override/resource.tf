resource "statusdrift_on_call_schedule_override" "holiday" {
  schedule_id           = statusdrift_on_call_schedule.primary.id
  replacement_member_id = 5
  starts_at             = "2024-12-24T00:00:00Z"
  ends_at               = "2024-12-26T00:00:00Z"
  note                  = "Holiday coverage"
}
