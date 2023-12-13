# rest api gateway for lambda functions /api/pulls/{id}

resource "aws_api_gateway_rest_api" "api" {
  name = "api"
}

resource "aws_api_gateway_resource" "pulls" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "pulls"
}

resource "aws_api_gateway_resource" "pull" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.pulls.id
  path_part   = "{id}"
}

resource "aws_api_gateway_method" "proxy" {
  authorization = "NONE"
  http_method   = "ANY"
  resource_id   = aws_api_gateway_resource.pull.id
  rest_api_id   = aws_api_gateway_rest_api.api.id
}

resource "aws_lambda_function" "get_pull" {
  function_name = "get-pull"
  role          = aws_iam_role.lambda.arn
  handler       = "get-pull"
  memory_size   = 128
  architectures = ["arm64"]
  timeout       = 30

  s3_key            = data.aws_s3_object.lambda_zip["get-pull"].key
  s3_object_version = data.aws_s3_object.lambda_zip["get-pull"].version_id
  s3_bucket         = aws_s3_bucket.corecheck-lambdas.id

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

resource "aws_api_gateway_integration" "lambda" {
  http_method             = aws_api_gateway_method.proxy.http_method
  resource_id             = aws_api_gateway_resource.pull.id
  rest_api_id             = aws_api_gateway_rest_api.api.id
  integration_http_method = "GET"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.get_pull.invoke_arn
}

resource "aws_api_gateway_deployment" "api" {
  depends_on = [
    aws_api_gateway_integration.lambda,
  ]
  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = "prod"
}

