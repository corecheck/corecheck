locals {
  dashboard_stack_name           = "corecheck-dashboard-${terraform.workspace}"
  dashboard_cloudwatch_namespace = terraform.workspace == "default" ? "Corecheck/prod" : "Corecheck/${terraform.workspace}"
}

resource "aws_grafana_workspace" "dashboard" {
  name = local.dashboard_stack_name

  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["AWS_SSO"]
  data_sources             = ["CLOUDWATCH"]
  description              = "Corecheck dashboard workspace for ${terraform.workspace}"
  permission_type          = "CUSTOMER_MANAGED"
  role_arn                 = aws_iam_role.dashboard_grafana.arn

  depends_on = [
    aws_iam_role_policy_attachment.dashboard_grafana_cloudwatch,
  ]
}
