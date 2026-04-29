locals {
  dashboard_stack_name = "corecheck-dashboard-${terraform.workspace}"
  dashboard_timestream_tables = {
    dashboard_metrics = {
      memory_store_retention_period_in_hours  = 24
      magnetic_store_retention_period_in_days = 365
    }
    dashboard_rollups = {
      memory_store_retention_period_in_hours  = 24
      magnetic_store_retention_period_in_days = 365
    }
  }
}

resource "aws_timestreamwrite_database" "dashboard" {
  provider      = aws.compute_region
  database_name = local.dashboard_stack_name
}

resource "aws_timestreamwrite_table" "dashboard" {
  provider = aws.compute_region
  for_each = local.dashboard_timestream_tables

  database_name = aws_timestreamwrite_database.dashboard.database_name
  table_name    = each.key

  retention_properties {
    memory_store_retention_period_in_hours  = each.value.memory_store_retention_period_in_hours
    magnetic_store_retention_period_in_days = each.value.magnetic_store_retention_period_in_days
  }
}

resource "aws_grafana_workspace" "dashboard" {
  name = local.dashboard_stack_name

  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["AWS_SSO"]
  data_sources             = ["CLOUDWATCH", "TIMESTREAM"]
  description              = "Corecheck dashboard workspace for ${terraform.workspace}"
  permission_type          = "CUSTOMER_MANAGED"
  role_arn                 = aws_iam_role.dashboard_grafana.arn

  depends_on = [
    aws_iam_role_policy_attachment.dashboard_grafana_cloudwatch,
    aws_iam_role_policy_attachment.dashboard_grafana_timestream,
  ]
}
