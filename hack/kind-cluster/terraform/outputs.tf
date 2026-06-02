output "ingress_class_name" {
  value       = var.ingress_class_name
  description = "IngressClass name used by ingress objects."
}

output "tls_secret_ref" {
  value       = "${var.tls_namespace}/${var.tls_secret_name}"
  description = "Namespace/name of the TLS secret intended for ingresses."
}

output "grafana_url" {
  value       = var.enable_monitoring ? "https://${local.grafana_host}:30443/" : "monitoring disabled"
  description = "Grafana UI URL (via HAProxy ingress NodePort)."
}

output "grafana_admin_user" {
  value       = var.enable_monitoring ? var.grafana_admin_user : "monitoring disabled"
  description = "Grafana admin username."
}

output "grafana_admin_password" {
  value       = var.enable_monitoring ? nonsensitive(var.grafana_admin_password) : "monitoring disabled"
  description = "Grafana admin password."
}

output "argocd_url" {
  value       = "https://argocd.localhost:30443/"
  description = "Argo CD UI URL (via HAProxy ingress NodePort)."
}

output "argocd_admin_password" {
  value       = data.kubernetes_secret_v1.argocd_initial_admin.data["password"]
  description = "Initial Argo CD admin password (also written to .state/argocd.env)."
  sensitive   = true
}

output "argocd_env_file" {
  value       = local_file.argocd_env.filename
  description = "Path to .state/argocd.env with Argo CD URL and admin password."
}

output "mq_web_console_url" {
  value       = "https://${local.mq_host}:30443/ibmmq/console/"
  description = "IBM MQ web console URL (via HAProxy ingress NodePort)."
}

output "mq_rest_admin_url" {
  value       = "https://${local.mq_host}:30443/ibmmq/rest/v2/admin/qmgr"
  description = "IBM MQ administrative REST API base (via HAProxy ingress NodePort)."
}

output "mq_in_cluster_endpoint" {
  value       = "https://ibm-mq.${var.mq_namespace}.svc:9443"
  description = "In-cluster mqweb endpoint for the operator's QueueManagerConnection.endpoint."
}

output "mq_admin_user" {
  value       = "admin"
  description = "IBM MQ admin user (REST admin API / MQWebAdmin role)."
}

output "mq_admin_password" {
  value       = nonsensitive(var.mq_admin_password)
  description = "IBM MQ admin user password."
}
