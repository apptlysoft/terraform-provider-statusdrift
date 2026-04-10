# Changelog

## 1.1.0

FEATURES:

* **Resource `statusdrift_sla_policy`:** Add `scope` attribute for organization-wide or group-specific SLA policies
* **Resource `statusdrift_monitor`:** Add `assertion` block for response body validation using JSON Path, XPath, or regex expressions with configurable comparison operators
* **Resource `statusdrift_on_call_schedule`:** Add `scope_type` and `monitor_group_id` attributes for group-scoped on-call schedules

## 1.0.0

FEATURES:

* **New Resource:** `statusdrift_monitor` — HTTP, HTTPS, TCP, UDP, PING, DNS, and PUSH monitors with location selection and embedded alerts
* **New Resource:** `statusdrift_monitor_group` — monitor grouping
* **New Resource:** `statusdrift_tag` — resource tagging with color support
* **New Resource:** `statusdrift_notification_channel` — 20 channel types (email, Slack, PagerDuty, OpsGenie, Datadog, and more)
* **New Resource:** `statusdrift_alert` — standalone alert rule management
* **New Resource:** `statusdrift_status_page` — public and private status pages with themes, social links, and custom CSS
* **New Resource:** `statusdrift_status_page_section` — status page sections
* **New Resource:** `statusdrift_status_page_announcement` — status page announcements
* **New Resource:** `statusdrift_maintenance_window` — one-time, daily, weekly, and monthly maintenance windows
* **New Resource:** `statusdrift_incident` — manual incident creation and resolution
* **New Resource:** `statusdrift_api_key` — API key management with granular permissions, expiration, and sensitive key handling
* **New Resource:** `statusdrift_sla_policy` — SLA policies with uptime targets, window types, response time SLA, and monitor/tag/group selection
* **New Resource:** `statusdrift_on_call_schedule` — on-call rotation scheduling with daily/weekly frequency and timezone support
* **New Resource:** `statusdrift_on_call_schedule_override` — on-call schedule overrides for temporary coverage
* **New Resource:** `statusdrift_on_call_policy` — on-call escalation policies with multi-step escalation, schedule and member targets
* **New Data Source:** Singular and plural data sources for all resource types (lookup by ID/name, or list all)
* **New Data Source:** `statusdrift_locations` — available monitoring regions
* **New Data Source:** `statusdrift_collaborator` / `statusdrift_collaborators` — lookup and list organization collaborators
