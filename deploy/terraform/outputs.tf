output "dashboard_cloudwatch_namespace" {
  description = "CloudWatch namespace used for dashboard telemetry metrics"
  value       = module.monitoring.dashboard_cloudwatch_namespace
}

output "public_dashboard_template_catalog" {
  description = "Repo-managed importable Grafana dashboard templates for the four public pages"
  value       = module.monitoring.public_dashboard_template_catalog
}

output "public_grafana_domain" {
  description = "Public hostname for the self-hosted Grafana OSS deployment"
  value       = module.monitoring.public_grafana_domain
}

output "public_grafana_base_url" {
  description = "Base URL for the self-hosted Grafana OSS deployment"
  value       = module.monitoring.public_grafana_base_url
}

output "public_grafana_dashboard_urls" {
  description = "Self-hosted Grafana URLs for the four public dashboard pages"
  value       = module.monitoring.public_grafana_dashboard_urls
}

output "public_dashboard_env_overrides" {
  description = "PUBLIC_DASHBOARD_* environment variable values pointing at the self-hosted Grafana dashboards"
  value       = module.monitoring.public_dashboard_env_overrides
}
