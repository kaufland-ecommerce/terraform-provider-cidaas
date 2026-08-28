resource "cidaas_notification_service_setup" "ses_email" {
  name                  = "SES Email"
  service_id            = "custom-ses-email"
  communication_methods = ["email"]
}

resource "cidaas_notification_provider_config" "ses_email" {
  service_setup_id = cidaas_notification_service_setup.ses_email.id

  config_data = jsonencode({
    commProvider = "custom-ses-email"
    commMethod   = "email"
    schemaData = {
      accessKeyId     = var.ses_access_key_id
      secretAccessKey = var.ses_secret_access_key
      region          = "eu-central-1"
    }
  })
}
