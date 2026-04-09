# StatusDrift Terraform Provider

The StatusDrift provider manages [StatusDrift](https://statusdrift.com) uptime monitoring infrastructure as code.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for building from source)

## Usage

```hcl
terraform {
  required_providers {
    statusdrift = {
      source = "apptlysoft/statusdrift"
    }
  }
}

provider "statusdrift" {
  api_key = var.statusdrift_api_key # Or set STATUSDRIFT_API_KEY env var
}
```

## Authentication

The provider requires an API key with the `sdk_` prefix. Set it via:

- Provider configuration: `api_key = "sdk_..."`
- Environment variable: `STATUSDRIFT_API_KEY`

Generate API keys in the StatusDrift dashboard or manage them with the `statusdrift_api_key` resource.

## Resources

| Resource | Description |
|---|---|
| `statusdrift_monitor` | HTTP, TCP, UDP, PING, DNS, and PUSH monitors |
| `statusdrift_monitor_group` | Monitor grouping |
| `statusdrift_tag` | Resource tagging |
| `statusdrift_notification_channel` | Email, Slack, PagerDuty, and 17 more channel types |
| `statusdrift_alert` | Alert rules for monitors |
| `statusdrift_status_page` | Public and private status pages |
| `statusdrift_status_page_section` | Status page sections |
| `statusdrift_status_page_announcement` | Status page announcements |
| `statusdrift_maintenance_window` | Scheduled maintenance windows |
| `statusdrift_incident` | Manual incident management |
| `statusdrift_api_key` | API key management with granular permissions |
| `statusdrift_sla_policy` | SLA policies with uptime targets |
| `statusdrift_on_call_schedule` | On-call rotation scheduling |
| `statusdrift_on_call_schedule_override` | On-call schedule overrides |
| `statusdrift_on_call_policy` | On-call escalation policies |

## Data Sources

Each resource has a singular data source (lookup by ID or name) and a plural data source (list all). Additional data sources include `statusdrift_locations` for available monitoring regions and `statusdrift_collaborator` / `statusdrift_collaborators` for looking up organization members.

## Development

```bash
# Build
make build

# Install locally
make install

# Run tests
make test

# Generate documentation
make generate

# Cross-compile for all platforms
make build-all
```

## License

Mozilla Public License 2.0 — see [LICENSE](LICENSE).
