data "statusdrift_sla_policy" "by_name" {
  name = "Production SLA"
}

data "statusdrift_sla_policy" "by_id" {
  id = 1
}
