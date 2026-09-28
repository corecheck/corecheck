resource "aws_cloudwatch_log_group" "github_events" {
  provider          = aws.compute_region
  name              = "/corecheck/github-events/${terraform.workspace}"
  retention_in_days = var.github_events_log_retention_days
}

resource "aws_ssm_parameter" "github_events_last_run" {
  provider = aws.compute_region
  name     = "/corecheck/${terraform.workspace}/github-events-last-run"
  type     = "String"
  value    = "initial"

  lifecycle {
    # The lambda updates this value at runtime; ignore post-creation changes.
    ignore_changes = [value]
  }
}

data "aws_iam_policy_document" "allow_stats_cloudwatch" {
  statement {
    effect = "Allow"
    actions = [
      "ssm:GetParameter",
      "ssm:PutParameter",
    ]
    resources = [aws_ssm_parameter.github_events_last_run.arn]
  }

  statement {
    effect = "Allow"
    actions = [
      "cloudwatch:PutMetricData",
    ]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "stats_cloudwatch_policy" {
  name        = "AllowStatsCloudWatchPolicy-${terraform.workspace}"
  description = "Allow the stats lambda to store GitHub events and open-count gauges"
  policy      = data.aws_iam_policy_document.allow_stats_cloudwatch.json
}

resource "aws_iam_role_policy_attachment" "stats_cloudwatch_policy_attachment" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.stats_cloudwatch_policy.arn
}
