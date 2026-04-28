output "dashboard_grafana_workspace_endpoint" {
  description = "Amazon Managed Grafana endpoint for the dashboard workspace"
  value       = aws_grafana_workspace.dashboard.endpoint
}

output "dashboard_grafana_workspace_id" {
  description = "Amazon Managed Grafana workspace ID for the dashboard workspace"
  value       = aws_grafana_workspace.dashboard.id
}

output "dashboard_timestream_database_name" {
  description = "Timestream database name for dashboard metrics"
  value       = aws_timestreamwrite_database.dashboard.database_name
}

output "dashboard_timestream_table_names" {
  description = "Initial Timestream table names for dashboard metrics"
  value       = { for name, table in aws_timestreamwrite_table.dashboard : name => table.table_name }
}

output "public_dashboard_template_catalog" {
  description = "Repo-managed importable Grafana dashboard templates for the four public pages"
  value = {
    for key, dashboard in local.public_dashboard_templates : key => {
      route                   = dashboard.route
      title                   = local.rendered_public_dashboard_templates[key].title
      grafana_uid             = local.rendered_public_dashboard_templates[key].uid
      public_grafana_url      = local.public_grafana_dashboard_urls[key]
      datadog_dashboard_id    = dashboard.datadog_dashboard_id
      datadog_dashboard_title = dashboard.datadog_dashboard_title
      datadog_widget_count    = dashboard.datadog_widget_count
      public_dashboard_env    = dashboard.public_dashboard_env
      template_file           = dashboard.template_file
    }
  }
}

output "public_grafana_domain" {
  description = "Public hostname for the self-hosted Grafana OSS deployment"
  value       = local.public_grafana_domain
}

output "public_grafana_base_url" {
  description = "Base URL for the self-hosted Grafana OSS deployment"
  value       = local.public_grafana_base_url
}

output "public_grafana_dashboard_urls" {
  description = "Self-hosted Grafana URLs for the four public dashboard pages"
  value       = local.public_grafana_dashboard_urls
}

output "public_dashboard_env_overrides" {
  description = "PUBLIC_DASHBOARD_* environment variable values pointing at the self-hosted Grafana dashboards"
  value       = local.public_dashboard_env_overrides
}
