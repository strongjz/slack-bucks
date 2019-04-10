
resource "aws_api_gateway_rest_api" "buck" {
  name = "Serverlessbuck"
  description = "Terraform Serverless Application buck"
}

resource "aws_api_gateway_resource" "buck" {
  rest_api_id = "${aws_api_gateway_rest_api.buck.id}"
  parent_id = "${aws_api_gateway_rest_api.buck.root_resource_id}"
  path_part   = "{proxy+}"
}

resource "aws_api_gateway_method" "buck" {
  rest_api_id = "${aws_api_gateway_rest_api.buck.id}"
  resource_id = "${aws_api_gateway_resource.buck.id}"

  http_method   = "ANY"
  authorization = "NONE"
  request_parameters = {
    "method.request.path.proxy" = true
  }
}

resource "aws_api_gateway_integration" "lambda" {
  rest_api_id = "${aws_api_gateway_rest_api.buck.id}"
  resource_id = "${aws_api_gateway_method.buck.resource_id}"
  http_method = "${aws_api_gateway_method.buck.http_method}"

  integration_http_method = "ANY"

  request_parameters =  {
    "integration.request.path.proxy" = "method.request.path.proxy"
  }

  type = "AWS_PROXY"
  uri = "${aws_lambda_function.buck.invoke_arn}"
}

resource "aws_api_gateway_method" "proxy_root" {
  rest_api_id = "${aws_api_gateway_rest_api.buck.id}"
  resource_id = "${aws_api_gateway_rest_api.buck.root_resource_id}"
  http_method = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_root" {
  rest_api_id = "${aws_api_gateway_rest_api.buck.id}"
  resource_id = "${aws_api_gateway_method.proxy_root.resource_id}"
  http_method = "${aws_api_gateway_method.proxy_root.http_method}"

  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "${aws_lambda_function.buck.invoke_arn}"
}


resource "aws_api_gateway_deployment" "buck" {
  depends_on = [
    "aws_api_gateway_integration.lambda",
    "aws_api_gateway_integration.lambda_root",
  ]

  rest_api_id = "${aws_api_gateway_rest_api.buck.id}"
  stage_name = "buck"
}

resource "aws_lambda_permission" "apigw" {
  statement_id = "AllowAPIGatewayInvoke"
  action = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.buck.arn}"
  principal = "apigateway.amazonaws.com"

  # The /*/* portion grants access from any method on any resource
  # within the API Gateway "REST API".
  source_arn = "${aws_api_gateway_deployment.buck.execution_arn}/*/*"
}
