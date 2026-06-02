# Shared mkcert TLS secret in the ingress controller namespace (created by the HAProxy chart).
resource "kubernetes_secret_v1" "wildcard_localhost_tls" {
  metadata {
    name      = var.tls_secret_name
    namespace = var.tls_namespace
  }

  type = "kubernetes.io/tls"

  data = {
    "tls.crt" = base64decode(var.tls_cert_string)
    "tls.key" = base64decode(var.tls_key_string)
  }

  depends_on = [helm_release.haproxy_ingress]
}
