provider "cidaas" {
  host          = var.cidaas_host          # optionally use CIDAAS_HOST env var
  client_id     = var.cidaas_client_id     # optionally use CIDAAS_CLIENT_ID env var
  client_secret = var.cidaas_client_secret # optionally use CIDAAS_CLIENT_SECRET env var
}