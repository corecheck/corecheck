data "aws_iam_policy_document" "canary_assume_role" {
  count = local.canary_enabled ? 1 : 0

  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "canary" {
  count              = local.canary_enabled ? 1 : 0
  name               = "corecheck-canary-${terraform.workspace}"
  assume_role_policy = data.aws_iam_policy_document.canary_assume_role[0].json
}

data "aws_iam_policy_document" "canary" {
  count = local.canary_enabled ? 1 : 0

  statement {
    actions   = ["s3:PutObject", "s3:GetObject"]
    resources = ["${aws_s3_bucket.canary_artifacts[0].arn}/*"]
  }

  statement {
    actions   = ["s3:GetBucketLocation"]
    resources = [aws_s3_bucket.canary_artifacts[0].arn]
  }

  statement {
    actions   = ["s3:ListAllMyBuckets"]
    resources = ["*"]
  }

  statement {
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = ["arn:aws:logs:*:*:*"]
  }

  statement {
    actions   = ["cloudwatch:PutMetricData"]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "cloudwatch:namespace"
      values   = ["CloudWatchSynthetics"]
    }
  }

  statement {
    actions   = ["xray:PutTraceSegments"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "canary" {
  count  = local.canary_enabled ? 1 : 0
  name   = "corecheck-canary-${terraform.workspace}"
  policy = data.aws_iam_policy_document.canary[0].json
}

resource "aws_iam_role_policy_attachment" "canary" {
  count      = local.canary_enabled ? 1 : 0
  role       = aws_iam_role.canary[0].name
  policy_arn = aws_iam_policy.canary[0].arn
}

# Telegram notifier Lambda IAM (only created when Telegram is configured)

data "aws_iam_policy_document" "telegram_lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "telegram_lambda" {
  count              = var.telegram_bot_token != "" ? 1 : 0
  name               = "corecheck-telegram-notifier-${terraform.workspace}"
  assume_role_policy = data.aws_iam_policy_document.telegram_lambda_assume_role.json
}

resource "aws_iam_role_policy_attachment" "telegram_lambda_basic" {
  count      = var.telegram_bot_token != "" ? 1 : 0
  role       = aws_iam_role.telegram_lambda[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

data "aws_iam_policy_document" "dashboard_grafana_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["grafana.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "dashboard_grafana" {
  name               = "${local.dashboard_stack_name}-grafana"
  assume_role_policy = data.aws_iam_policy_document.dashboard_grafana_assume_role.json
}

resource "aws_iam_role_policy_attachment" "dashboard_grafana_cloudwatch" {
  role       = aws_iam_role.dashboard_grafana.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonGrafanaCloudWatchAccess"
}
