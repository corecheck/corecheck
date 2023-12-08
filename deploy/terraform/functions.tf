locals {
  lambdas = [
    "github-sync",
    "migrate",
  ]

  # create a map of lambdas and their environment variables
  lambda_env = {
    "github-sync" = {
      DATABASE_HOST     = aws_instance.db.public_ip
      DATABASE_PORT     = 5432
      DATABASE_USER     = var.db_user
      DATABASE_PASSWORD = var.db_password
      DATABASE_NAME     = var.db_database

      SQS_QUEUE_URL = aws_sqs_queue.corecheck_queue.url

      GITHUB_ACCESS_TOKEN = var.github_token
    },
    "migrate" = {
      DATABASE_HOST     = aws_instance.db.public_ip
      DATABASE_PORT     = 5432
      DATABASE_USER     = var.db_user
      DATABASE_PASSWORD = var.db_password
      DATABASE_NAME     = var.db_database
    }
  }
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

# AWSLambdaVPCAccessExecutionRole
data "aws_iam_policy" "lambda_vpc_access" {
  arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

# data "aws_s3_object" "lambda_zip" {
#   for_each = toset(local.lambdas)
#   bucket   = aws_s3_bucket.corecheck-lambdas.id
#   key      = "${each.value}.zip"
# }

# resource "aws_cloudwatch_log_group" "function_logs" {
#   for_each = toset(local.lambdas)
#   name     = "/aws/lambda/${each.value}"

#   retention_in_days = 7

#   lifecycle {
#     create_before_destroy = true
#     prevent_destroy       = false
#   }
# }

# resource "aws_lambda_function" "function" {
#   function_name = "migrate"
#   handler       = "migrate"
#   description   = "Syncs github repositories with the database"
#   role          = aws_iam_role.lambda.arn
#   memory_size   = 128
#   architectures = ["arm64"]
#   timeout       = 60

#   s3_key            = data.aws_s3_object.lambda_zip["migrate"].key
#   s3_object_version = data.aws_s3_object.lambda_zip["migrate"].version_id
#   s3_bucket         = aws_s3_bucket.corecheck-lambdas.id

#   environment {
#     variables = local.lambda_env["migrate"]
#   }

#   runtime = "provided.al2"
# }


# resource "aws_lambda_invocation" "invoke" {
#   function_name = "migrate"
#   input = "{\"action\": \"up\"}"
#   depends_on = [
#     aws_lambda_function.function,
#   ]
# }
