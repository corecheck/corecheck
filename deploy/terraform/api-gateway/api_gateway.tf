resource "aws_api_gateway_rest_api" "api" {
  name = "api-${terraform.workspace}"
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

resource "aws_lambda_permission" "api_gw" {
  for_each      = toset(local.api_lambdas)
  function_name = "${each.value}-${terraform.workspace}"
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.api.execution_arn}/*"

  depends_on = [
    aws_api_gateway_deployment.api,
    aws_lambda_function.lambda,
  ]
}

resource "aws_api_gateway_integration" "lambda" {
  http_method             = aws_api_gateway_method.get_pull.http_method
  resource_id             = aws_api_gateway_resource.get_pull.id
  rest_api_id             = aws_api_gateway_rest_api.api.id
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.lambda["get-pull"].invoke_arn
}

resource "aws_api_gateway_integration" "lambda_list" {
  http_method             = aws_api_gateway_method.list_pulls.http_method
  resource_id             = aws_api_gateway_resource.pulls.id
  rest_api_id             = aws_api_gateway_rest_api.api.id
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.lambda["list-pulls"].invoke_arn
}

resource "aws_api_gateway_deployment" "api" {
  rest_api_id       = aws_api_gateway_rest_api.api.id
  stage_name        = "api"
  description       = md5(file("api-gateway/api_gateway.tf"))
  stage_description = md5(file("api-gateway/api_gateway.tf"))
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

resource "aws_cloudwatch_log_group" "api_gateway_logs" {
  name = "/aws/api-gateway/${aws_api_gateway_rest_api.api.id}"
  retention_in_days = 7
}
