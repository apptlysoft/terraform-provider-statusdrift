data "statusdrift_api_key" "by_name" {
  name = "CI/CD Pipeline Key"
}

data "statusdrift_api_key" "by_id" {
  id = 42
}
