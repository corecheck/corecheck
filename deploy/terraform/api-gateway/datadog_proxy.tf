resource "aws_api_gateway_rest_api" "datadog_proxy" {
  name = "datadog-proxy-${terraform.workspace}"
}

resource "aws_api_gateway_resource" "datadog_proxy" {
  rest_api_id = aws_api_gateway_rest_api.datadog_proxy.id
  parent_id   = aws_api_gateway_rest_api.datadog_proxy.root_resource_id
  path_part   = "{proxy+}"
}

resource "aws_api_gateway_method" "datadog_proxy" {
  authorization = "NONE"
  http_method   = "ANY"
  resource_id   = aws_api_gateway_resource.datadog_proxy.id
  rest_api_id   = aws_api_gateway_rest_api.datadog_proxy.id
}

resource "aws_lambda_permission" "datadog_proxy" {
  function_name = "datadog-proxy-${terraform.workspace}"
  statement_id  = "AllowAPIGatewayInvokeCorecheckDatadogProxy"
  action        = "lambda:InvokeFunction"
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.datadog_proxy.execution_arn}/*"

  depends_on = [
    aws_api_gateway_deployment.datadog_proxy,
    aws_lambda_function.lambda
  ]
}

resource "aws_api_gateway_integration" "datadog_proxy" {
  http_method             = aws_api_gateway_method.datadog_proxy.http_method
  resource_id             = aws_api_gateway_resource.datadog_proxy.id
  rest_api_id             = aws_api_gateway_rest_api.datadog_proxy.id
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.lambda["datadog-proxy"].invoke_arn
}

resource "aws_api_gateway_deployment" "datadog_proxy" {
  rest_api_id       = aws_api_gateway_rest_api.datadog_proxy.id
  stage_name        = "datadog-proxy"
  description       = md5(file("api-gateway/datadog_proxy.tf"))
  stage_description = md5(file("api-gateway/datadog_proxy.tf"))
  lifecycle {
    create_before_destroy = true
    prevent_destroy       = false
  }
  depends_on = [
    aws_api_gateway_method.datadog_proxy,
    aws_api_gateway_integration.datadog_proxy,
  ]
}

resource "aws_cloudwatch_log_group" "datadog_proxy_logs" {
  name = "/aws/datadog-proxy/${aws_api_gateway_rest_api.datadog_proxy.name}-${terraform.workspace}"
  retention_in_days = 7
}
