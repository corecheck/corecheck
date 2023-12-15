locals {
  api_lambdas = [
    "get-pull",
    "list-pulls",
  ]
}

data "aws_s3_object" "lambda_api_zip" {
  for_each = toset(local.api_lambdas)
  bucket   = var.s3_bucket
  key      = "${each.value}.zip"
}

resource "aws_lambda_function" "lambda" {
  for_each      = toset(local.api_lambdas)
  function_name = each.value
  handler       = each.value
  role          = aws_iam_role.lambda.arn
  memory_size   = 128
  architectures = ["arm64"]
  timeout       = 30

  s3_key            = data.aws_s3_object.lambda_api_zip[each.value].key
  s3_object_version = data.aws_s3_object.lambda_api_zip[each.value].version_id
  s3_bucket         = var.s3_bucket
  environment {
    variables = {
      DATABASE_HOST     = var.db_host
      DATABASE_PORT     = var.db_port
      DATABASE_USER     = var.db_user
      DATABASE_PASSWORD = var.db_password
      DATABASE_NAME     = var.db_database
    }
  }

  runtime = "provided.al2"
}

resource "aws_cloudwatch_log_group" "function_api_logs" {
  for_each = toset(local.api_lambdas)
  name     = "/aws/lambda/${each.value}-${terraform.workspace}"

  retention_in_days = 7

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = false
  }
}
