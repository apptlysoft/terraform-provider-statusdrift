resource "statusdrift_api_key" "cicd" {
  name            = "CI/CD Pipeline Key"
  expires_in_days = 180

  permissions = {
    monitor = {
      create = true
      read   = true
      update = true
      delete = false
    }
    tag = {
      read = true
    }
    status_page = {
      read = true
    }
  }
}

output "api_key_value" {
  value     = statusdrift_api_key.cicd.key
  sensitive = true
}
