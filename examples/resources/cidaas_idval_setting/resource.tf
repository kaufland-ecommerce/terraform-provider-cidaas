resource "cidaas_idval_setting" "age_check_de" {
  id                    = "b4ba52ae-87b3-4967-93b6-8e6ae8f39922"
  name                  = "Kaufland Marketplace Age Check DE"
  description           = "Kaufland Marketplace Age Check DE"
  mode                  = "AgeCheckEssential"
  allowed_redirect_uris = "https://account.kaufland.com https://www.kaufland.de https://kaufland.de"

  consent_config = {
    enabled = true
    consents = [{
      name      = "Consent"
      url       = "https://www.kaufland.de/i/rechtliches/datenschutz?hidebanner=true"
      mandatory = true
      localization = {
        de = "..."
        en = "..."
      }
    }]
  }
}
