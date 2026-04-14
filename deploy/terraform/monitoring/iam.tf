data "aws_iam_policy_document" "canary_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "canary" {
  name               = "corecheck-canary-${terraform.workspace}"
  assume_role_policy = data.aws_iam_policy_document.canary_assume_role.json
}

data "aws_iam_policy_document" "canary" {
  statement {
    actions   = ["s3:PutObject", "s3:GetObject"]
    resources = ["${aws_s3_bucket.canary_artifacts.arn}/*"]
  }

  statement {
    actions   = ["s3:GetBucketLocation"]
    resources = [aws_s3_bucket.canary_artifacts.arn]
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
  name   = "corecheck-canary-${terraform.workspace}"
  policy = data.aws_iam_policy_document.canary.json
}

resource "aws_iam_role_policy_attachment" "canary" {
  role       = aws_iam_role.canary.name
  policy_arn = aws_iam_policy.canary.arn
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
