locals {
  lambdas = [
    "github-sync",
    "migrate",
    "handle-coverage",
    "handle-benchmarks",
    "get-pull",
  ]
}

data "aws_iam_policy_document" "assume_lambda_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

// create lambda role, that lambda function can assume (use)
resource "aws_iam_role" "lambda" {
  name               = "AssumeLambdaRole"
  description        = "Role for lambda to assume lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role.json
}


data "aws_iam_policy_document" "allow_lambda_logging" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = [
      "arn:aws:logs:*:*:*",
    ]
  }
}

data "aws_iam_policy_document" "allow_lambda_sqs" {
  statement {
    effect = "Allow"
    actions = [
      "sqs:SendMessage",
    ]

    resources = [
      aws_sqs_queue.corecheck_queue.arn,
    ]
  }
}

resource "aws_iam_policy" "function_logging_policy" {
  name        = "AllowLambdaLoggingPolicy"
  description = "Policy for lambda cloudwatch logging"
  policy      = data.aws_iam_policy_document.allow_lambda_logging.json
}

// attach policy to out created lambda role
resource "aws_iam_role_policy_attachment" "lambda_logging_policy_attachment" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.function_logging_policy.arn
}

data "aws_s3_object" "lambda_zip" {
  provider = aws.compute_region
  for_each = toset(local.lambdas)
  bucket   = aws_s3_bucket.corecheck-lambdas.id
  key      = "${each.value}.zip"
}

resource "aws_cloudwatch_log_group" "function_logs" {
  for_each = toset(local.lambdas)
  name     = "/aws/lambda/${each.value}"
  provider = aws.compute_region

  retention_in_days = 7

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = false
  }
}