locals {
  dashboard_stack_name           = "corecheck-dashboard-${terraform.workspace}"
  dashboard_cloudwatch_namespace = terraform.workspace == "default" ? "Corecheck/prod" : "Corecheck/${terraform.workspace}"
}
