resource "cidaas_hook" "example_hook" {
  url = "https://example.com"
  events = [
    "ACCOUNT_MODIFIED",
    "APP_MODIFIED",
    "LOGIN_WITH_SOCIAL",
    "LOGIN_WITH_CIDAAS",
    "LOGIN_FAILURE"
  ]
  auth_type = "APIKEY"
  apikey_details = {
    apikey             = "keyboardcat"
    apikey_placeholder = "X-API-Key"
    apikey_placement   = "header"
  }
}