data "statusdrift_sla_policies" "all" {}

output "all_sla_policy_names" {
  value = [for p in data.statusdrift_sla_policies.all.sla_policies : p.name]
}
