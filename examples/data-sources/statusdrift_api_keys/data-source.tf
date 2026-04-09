data "statusdrift_api_keys" "all" {}

output "all_api_key_names" {
  value = [for k in data.statusdrift_api_keys.all.api_keys : k.name]
}
