resource "cidaas_theme" "idval_kaufland" {
  name = "idvalKaufland"
  css  = file("${path.module}/idval-kaufland.css")
}
