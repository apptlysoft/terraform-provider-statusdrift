terraform {
  required_providers {
    statusdrift = {
      source = "apptlysoft/statusdrift"
    }
  }
}

provider "statusdrift" {
  api_url = "https://api.statusdrift.com" # Optional, defaults to this
  api_key = var.statusdrift_api_key        # Or set STATUSDRIFT_API_KEY env var
}

variable "statusdrift_api_key" {
  type      = string
  sensitive = true
}
