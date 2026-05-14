resource "aws_cloudwatch_log_group" "github_events" {
  provider          = aws.compute_region
  name              = "/corecheck/github-events/${terraform.workspace}"
  retention_in_days = var.github_events_log_retention_days
}

resource "aws_ssm_parameter" "github_events_git_sha" {
  provider = aws.compute_region
  name     = "/corecheck/${terraform.workspace}/github-events-git-sha"
  type     = "String"
  value    = ""

  lifecycle {
    # The lambda updates this value at runtime; ignore post-creation changes.
    ignore_changes = [value]
  }
}

data "aws_iam_policy_document" "allow_stats_ssm" {
  statement {
    effect = "Allow"
    actions = [
      "ssm:GetParameter",
      "ssm:PutParameter",
    ]
    resources = [aws_ssm_parameter.github_events_git_sha.arn]
  }
}

resource "aws_iam_policy" "stats_ssm_policy" {
  name        = "AllowStatsSSMPolicy-${terraform.workspace}"
  description = "Allow the stats lambda to read/write the github-events git SHA in SSM"
  policy      = data.aws_iam_policy_document.allow_stats_ssm.json
}

resource "aws_iam_role_policy_attachment" "stats_ssm_policy_attachment" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.stats_ssm_policy.arn
}
