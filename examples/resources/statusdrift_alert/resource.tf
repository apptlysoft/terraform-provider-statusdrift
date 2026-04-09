# Alerts are managed through the statusdrift_monitor resource's alert block.
# Use the statusdrift_alert data source to read existing alerts.

data "statusdrift_monitor" "example" {
  id = 123
}
