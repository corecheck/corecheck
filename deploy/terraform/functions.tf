locals {
  binary_name  = "github-sync"
  src_path     = "../../functions/github-sync"
  binary_path  = "${local.src_path}/bootstrap"
  archive_path = "/tmp/${local.binary_name}.zip"
}

// allow lambda service to assume (use) the role with such policy
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

// create a policy to allow writing into logs and create logs stream
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

data "aws_s3_object" "function_zip" {
  bucket = aws_s3_bucket.corecheck-lambdas.id
  key    = "${local.binary_name}.zip"
  
}

// create the lambda function from zip file
resource "aws_lambda_function" "function" {
  function_name = "github-sync"
  description   = "Syncs github repositories with the database"
  role          = aws_iam_role.lambda.arn
  handler       = local.binary_name
  memory_size   = 128
  architectures = ["arm64"]
  timeout       = 60

  s3_key = data.aws_s3_object.function_zip.key
  s3_bucket = aws_s3_bucket.corecheck-lambdas.id
  s3_object_version = data.aws_s3_object.function_zip.version_id
  vpc_config {
    subnet_ids         = [
      data.aws_subnet_ids.example.ids[0],
      data.aws_subnet_ids.example.ids[1],
      data.aws_subnet_ids.example.ids[2],
    ]
    security_group_ids = [data.aws_security_group.default.id]
  }

  environment {
    variables = {
      DATABASE_HOST     = aws_instance.db.private_ip
      DATABASE_PORT     = 5432
      DATABASE_USER     = var.db_user
      DATABASE_PASSWORD = var.db_password
      DATABASE_NAME     = var.db_database

      SQS_QUEUE_URL = aws_sqs_queue.corecheck_queue.url

      GITHUB_ACCESS_TOKEN = var.github_token
    }
  }

  runtime = "provided.al2"
}

resource "aws_cloudwatch_log_group" "function_logs" {
  name = "/aws/lambda/${aws_lambda_function.function.function_name}"

  retention_in_days = 7

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = false
  }
}
