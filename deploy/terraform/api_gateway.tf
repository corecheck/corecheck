locals {
  api_lambdas = [
    "get-pull",
    "list-pulls",
  ]
}

data "aws_s3_object" "lambda_api_zip" {
  for_each = toset(local.api_lambdas)
  bucket   = aws_s3_bucket.corecheck-lambdas-api.id
  key      = "${each.value}.zip"
}

resource "aws_cloudwatch_log_group" "function_api_logs" {
  for_each = toset(local.api_lambdas)
  name     = "/aws/lambda/${each.value}"

  retention_in_days = 7

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = false
  }
}

resource "aws_api_gateway_rest_api" "api" {
  name = "api"
}

resource "aws_api_gateway_resource" "pulls" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "pulls"
}

resource "aws_api_gateway_resource" "get_pull" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.pulls.id
  path_part   = "{id}"
}

resource "aws_api_gateway_method" "get_pull" {
  authorization = "NONE"
  http_method   = "GET"
  resource_id   = aws_api_gateway_resource.get_pull.id
  rest_api_id   = aws_api_gateway_rest_api.api.id
}

resource "aws_api_gateway_method" "list_pulls" {
  authorization = "NONE"
  http_method   = "GET"
  resource_id   = aws_api_gateway_resource.pulls.id
  rest_api_id   = aws_api_gateway_rest_api.api.id
}

resource "aws_lambda_function" "get_pull" {
  function_name = "get-pull"
  role          = aws_iam_role.lambda.arn
  handler       = "get-pull"
  memory_size   = 128
  architectures = ["arm64"]
  timeout       = 30

  s3_key            = data.aws_s3_object.lambda_api_zip["get-pull"].key
  s3_object_version = data.aws_s3_object.lambda_api_zip["get-pull"].version_id
  s3_bucket         = aws_s3_bucket.corecheck-lambdas-api.id

  environment {
    variables = {
      DATABASE_HOST     = aws_instance.db.public_ip
      DATABASE_PORT     = 5432
      DATABASE_USER     = var.db_user
      DATABASE_PASSWORD = var.db_password
      DATABASE_NAME     = var.db_database
    }
  }

  runtime = "provided.al2"
}

resource "aws_lambda_function" "list_pulls" {
  function_name = "list-pulls"
  role          = aws_iam_role.lambda.arn
  handler       = "list-pulls"
  memory_size   = 128
  architectures = ["arm64"]
  timeout       = 30

  s3_key            = data.aws_s3_object.lambda_api_zip["list-pulls"].key
  s3_object_version = data.aws_s3_object.lambda_api_zip["list-pulls"].version_id
  s3_bucket         = aws_s3_bucket.corecheck-lambdas-api.id

  environment {
    variables = {
      DATABASE_HOST     = aws_instance.db.public_ip
      DATABASE_PORT     = 5432
      DATABASE_USER     = var.db_user
      DATABASE_PASSWORD = var.db_password
      DATABASE_NAME     = var.db_database
    }
  }

  runtime = "provided.al2"
}

resource "aws_lambda_permission" "api_gw" {
  for_each      = toset(local.api_lambdas)
  function_name = each.value
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.api.execution_arn}/*"
}


resource "aws_api_gateway_integration" "lambda" {
  http_method             = aws_api_gateway_method.get_pull.http_method
  resource_id             = aws_api_gateway_resource.get_pull.id
  rest_api_id             = aws_api_gateway_rest_api.api.id
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.get_pull.invoke_arn
}

resource "aws_api_gateway_integration" "lambda_list" {
  http_method             = aws_api_gateway_method.list_pulls.http_method
  resource_id             = aws_api_gateway_resource.pulls.id
  rest_api_id             = aws_api_gateway_rest_api.api.id
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.list_pulls.invoke_arn
}

# deployment
resource "aws_api_gateway_deployment" "api" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = "api"
  lifecycle {
    create_before_destroy = true
    prevent_destroy       = false
  }
  depends_on = [
    aws_api_gateway_method.get_pull,
    aws_api_gateway_method.list_pulls,
    aws_api_gateway_integration.lambda,
  ]
}
# api gateway log
resource "aws_cloudwatch_log_group" "api_gateway_logs" {
  name = "/aws/api-gateway/${aws_api_gateway_rest_api.api.id}"
}
