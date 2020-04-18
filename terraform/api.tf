resource "aws_api_gateway_rest_api" "app" {
  name        = "GoLangExample"
  description = "Terraform Serverless Application app"

}

resource "aws_api_gateway_resource" "app" {
  rest_api_id = aws_api_gateway_rest_api.app.id
  parent_id   = aws_api_gateway_rest_api.app.root_resource_id
  path_part   = "{proxy+}"
}

resource "aws_api_gateway_method" "app" {
  rest_api_id = aws_api_gateway_rest_api.app.id
  resource_id = aws_api_gateway_resource.app.id

  http_method   = "ANY"
  authorization = "NONE"

  request_parameters = {
    "method.request.path.proxy" = true
  }
}

resource "aws_api_gateway_integration" "lambda" {
  rest_api_id = aws_api_gateway_rest_api.app.id
  resource_id = aws_api_gateway_method.app.resource_id
  http_method = aws_api_gateway_method.app.http_method
  timeout_milliseconds = 29000

  integration_http_method = "POST"

  type = "AWS_PROXY"
  uri  = aws_lambda_function.app.invoke_arn
}


resource "aws_api_gateway_method" "proxy_root" {
  rest_api_id   = aws_api_gateway_rest_api.app.id
  resource_id   = aws_api_gateway_rest_api.app.root_resource_id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_root" {
  rest_api_id = aws_api_gateway_rest_api.app.id
  resource_id = aws_api_gateway_method.proxy_root.resource_id
  http_method = aws_api_gateway_method.proxy_root.http_method

  credentials = aws_lambda_permission.apigw_app.source_arn

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.app.invoke_arn
}

resource "aws_lambda_permission" "apigw_app" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.app.arn
  principal     = "apigateway.amazonaws.com"

  # "arn:aws:execute-api:region:account-id:api-id/stage/METHOD_HTTP_VERB/Resource-path"
  #https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-control-access-using-iam-policies-to-invoke-api.html#

  source_arn = "${aws_api_gateway_rest_api.app.execution_arn}/*/*/*"
}



resource "aws_api_gateway_deployment" "app" {
  depends_on = [
    aws_api_gateway_integration.lambda,
    aws_api_gateway_integration.lambda_root,
    aws_lambda_permission.apigw_app,
  ]

  rest_api_id = aws_api_gateway_rest_api.app.id
  stage_name  = "app"

  variables = {
    "deployed_at" = timestamp()
  }

  lifecycle {
    create_before_destroy = true
  }

}




