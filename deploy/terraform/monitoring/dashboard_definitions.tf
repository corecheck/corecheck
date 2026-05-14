data "aws_caller_identity" "current" {}

locals {
  public_dashboard_templates = {
    github = {
      route                   = "/"
      title                   = "Corecheck GitHub Overview"
      datadog_dashboard_id    = "b5p-ekj-pxv"
      datadog_dashboard_title = "Bitcoin Core GitHub Overview"
      datadog_widget_count    = 2
      public_dashboard_env    = "PUBLIC_DASHBOARD_GITHUB_URL"
      template_file           = "grafana-dashboard-templates/github-overview.json.tftpl"
    }
    tests = {
      route                   = "/tests"
      title                   = "Corecheck Tests"
      datadog_dashboard_id    = "7ck-zbu-au3"
      datadog_dashboard_title = "Bitcoin Core tests"
      datadog_widget_count    = 5
      public_dashboard_env    = "PUBLIC_DASHBOARD_TESTS_URL"
      template_file           = "grafana-dashboard-templates/tests.json.tftpl"
    }
    benchmarks = {
      route                   = "/benchmarks"
      title                   = "Corecheck Benchmarks"
      datadog_dashboard_id    = "qem-ga2-953"
      datadog_dashboard_title = "Bitcoin Core benchmarks"
      datadog_widget_count    = 6
      public_dashboard_env    = "PUBLIC_DASHBOARD_BENCHMARKS_URL"
      template_file           = "grafana-dashboard-templates/benchmarks.json.tftpl"
    }
    jobs = {
      route                   = "/jobs"
      title                   = "Corecheck Jobs"
      datadog_dashboard_id    = "sxt-cxy-nsc"
      datadog_dashboard_title = "Corecheck job executions"
      datadog_widget_count    = 35
      public_dashboard_env    = "PUBLIC_DASHBOARD_JOBS_URL"
      template_file           = "grafana-dashboard-templates/jobs.json.tftpl"
    }
  }

  dashboard_job_resources = {
    workflow_state_machine_arn = "arn:aws:states:${var.dashboard_compute_region}:${data.aws_caller_identity.current.account_id}:stateMachine:start-jobs-${terraform.workspace}"
    mutation_state_machine_arn = "arn:aws:states:${var.dashboard_compute_region}:${data.aws_caller_identity.current.account_id}:stateMachine:start-mutation-jobs-${terraform.workspace}"
    coverage_queue_name        = "coverage-queue-${terraform.workspace}"
    sonar_queue_name           = "sonar-queue-${terraform.workspace}"
    bench_queue_name           = "bench-queue-${terraform.workspace}"
    mutation_queue_name        = "mutation-queue-${terraform.workspace}"
  }

  public_dashboard_template_context = merge(local.dashboard_job_resources, {
    telemetry_namespace     = local.dashboard_cloudwatch_namespace
    compute_region          = var.dashboard_compute_region
    github_events_log_group = "/corecheck/github-events/${terraform.workspace}"
  })

  # Keep these as importable Grafana JSON templates for now. Provisioning workspace-local
  # dashboards directly from Terraform would require a separate Grafana auth/token lifecycle.
  rendered_public_dashboard_templates = {
    for key, dashboard in local.public_dashboard_templates : key => jsondecode(
      templatefile(
        "${path.module}/${dashboard.template_file}",
        merge(local.public_dashboard_template_context, {
          route                   = dashboard.route
          title                   = dashboard.title
          datadog_dashboard_id    = dashboard.datadog_dashboard_id
          datadog_dashboard_title = dashboard.datadog_dashboard_title
          datadog_widget_count    = dashboard.datadog_widget_count
          public_dashboard_env    = dashboard.public_dashboard_env
        })
      )
    )
  }
}

locals {
  public_grafana_name     = terraform.workspace == "default" ? "grafana" : "grafana-${terraform.workspace}"
  public_grafana_domain   = "${local.public_grafana_name}.${var.dns_name}"
  public_grafana_base_url = "https://${local.public_grafana_domain}"

  public_grafana_datasource_names = {
    cloudwatch = "Corecheck CloudWatch"
  }

  provisioned_public_dashboard_templates = {
    for key, dashboard in local.public_dashboard_templates : key => jsonencode({
      for k, v in jsondecode(replace(
        templatefile(
          "${path.module}/${dashboard.template_file}",
          merge(local.public_dashboard_template_context, {
            route                   = dashboard.route
            title                   = dashboard.title
            datadog_dashboard_id    = dashboard.datadog_dashboard_id
            datadog_dashboard_title = dashboard.datadog_dashboard_title
            datadog_widget_count    = dashboard.datadog_widget_count
            public_dashboard_env    = dashboard.public_dashboard_env
          })
        ),
        "\"$${DS_CLOUDWATCH}\"",
        jsonencode({
          type = "cloudwatch"
          uid  = "corecheck-cloudwatch"
        })
      )) : k => v if k != "__inputs"
    })
  }

  public_grafana_dashboard_urls = {
    for key, dashboard in local.rendered_public_dashboard_templates :
    key => "${local.public_grafana_base_url}/d/${dashboard.uid}/${dashboard.uid}?orgId=1&kiosk"
  }

  public_dashboard_env_overrides = {
    for key, dashboard in local.public_dashboard_templates :
    dashboard.public_dashboard_env => local.public_grafana_dashboard_urls[key]
  }
}
