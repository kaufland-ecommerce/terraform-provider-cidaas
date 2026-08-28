resource "cidaas_notification_service_setup" "ses_email" {
  name                  = "SES Email"
  service_id            = "custom-ses-email"
  communication_methods = ["email"]
}
