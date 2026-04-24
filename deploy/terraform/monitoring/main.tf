locals {
  canary_enabled = terraform.workspace != "dev"
}

# S3 bucket for canary artifacts (screenshots, logs, HAR files)
resource "aws_s3_bucket" "canary_artifacts" {
  count  = local.canary_enabled ? 1 : 0
  bucket = "corecheck-canary-artifacts-${terraform.workspace}"
}

resource "aws_s3_bucket_lifecycle_configuration" "canary_artifacts" {
  count  = local.canary_enabled ? 1 : 0
  bucket = aws_s3_bucket.canary_artifacts[0].id

  rule {
    id     = "expire-old-artifacts"
    status = "Enabled"
    expiration {
      days = 30
    }
  }
}

# Zip the canary script into the structure CloudWatch Synthetics expects:
# nodejs/node_modules/<handler-filename>.js
data "archive_file" "canary_script" {
  count       = local.canary_enabled ? 1 : 0
  type        = "zip"
  output_path = "${path.module}/dist/corecheck-health.zip"

  source {
    content  = file("${path.module}/canary_script/corecheck-health.js")
    filename = "nodejs/node_modules/corecheck-health.js"
  }
}

# Canary – runs every 30 minutes (configurable) and checks the full user journey
resource "aws_synthetics_canary" "corecheck_health" {
  count = local.canary_enabled ? 1 : 0

  # Name must be ≤21 characters and alphanumeric + hyphens only
  name                 = "cc-health-${terraform.workspace}"
  artifact_s3_location = "s3://${aws_s3_bucket.canary_artifacts[0].id}/artifacts"
  execution_role_arn   = aws_iam_role.canary[0].arn
  handler              = "corecheck-health.handler"
  zip_file             = data.archive_file.canary_script[0].output_path
  runtime_version      = "syn-nodejs-puppeteer-9.0"
  start_canary         = true

  schedule {
    expression = var.canary_schedule
  }

  depends_on = [aws_iam_role_policy_attachment.canary[0]]
}

# CloudWatch Alarm – fires after 6 hours of continuous canary failures (12 × 30-min windows).
# Uses SuccessPercent (always emitted on every run) rather than Failed (a COUNT that emits
# nothing on a passing run, which would cause treat_missing_data=breaching to keep the alarm
# stuck in ALARM permanently even after the canary recovers).
resource "aws_cloudwatch_metric_alarm" "canary_failed" {
  count               = local.canary_enabled ? 1 : 0
  alarm_name          = "corecheck-health-check-failed-${terraform.workspace}"
  alarm_description   = "CoreCheck health check canary has been failing for 6 hours. The frontend or coverage pipeline may be down."
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 12
  datapoints_to_alarm = 12
  metric_name         = "SuccessPercent"
  namespace           = "CloudWatchSynthetics"
  period              = 1800
  statistic           = "Average"
  threshold           = 100
  treat_missing_data  = "breaching"

  dimensions = {
    CanaryName = aws_synthetics_canary.corecheck_health[0].name
  }

  alarm_actions = [aws_sns_topic.alerts.arn]
  ok_actions    = [aws_sns_topic.alerts.arn]
}

# SNS topic for all corecheck alerts
resource "aws_sns_topic" "alerts" {
  name = "corecheck-alerts-${terraform.workspace}"
}

# Email subscription – AWS will send a confirmation email before activating
resource "aws_sns_topic_subscription" "email" {
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}

# --- Telegram notifications (optional) ---

data "archive_file" "telegram_lambda" {
  count       = var.telegram_bot_token != "" ? 1 : 0
  type        = "zip"
  output_path = "${path.module}/dist/telegram-notifier.zip"

  source {
    content  = file("${path.module}/telegram_lambda/index.js")
    filename = "index.js"
  }
}

resource "aws_lambda_function" "telegram_notifier" {
  count            = var.telegram_bot_token != "" ? 1 : 0
  function_name    = "corecheck-telegram-notifier-${terraform.workspace}"
  handler          = "index.handler"
  role             = aws_iam_role.telegram_lambda[0].arn
  runtime          = "nodejs22.x"
  filename         = data.archive_file.telegram_lambda[0].output_path
  source_code_hash = data.archive_file.telegram_lambda[0].output_base64sha256
  timeout          = 30

  environment {
    variables = {
      TELEGRAM_BOT_TOKEN = var.telegram_bot_token
      TELEGRAM_CHAT_ID   = var.telegram_chat_id
    }
  }
}

resource "aws_lambda_permission" "sns_invoke_telegram" {
  count         = var.telegram_bot_token != "" ? 1 : 0
  statement_id  = "AllowSNSInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.telegram_notifier[0].function_name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.alerts.arn
}

resource "aws_sns_topic_subscription" "telegram" {
  count     = var.telegram_bot_token != "" ? 1 : 0
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.telegram_notifier[0].arn
}
