resource "cidaas_theme" "idval_kaufland" {
  name = "idvalKaufland"
  css  = <<-CSS
    :root {
      --idval-primary-color: #2a2a2a;
    }
  CSS
}
