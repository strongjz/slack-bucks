resource "aws_api_gateway_rest_api" "buck" {
  name        = "Serverlessbuck"
  description = "Terraform Serverless Application buck"

}

resource "aws_api_gateway_resource" "buck" {
  rest_api_id = "${aws_api_gateway_rest_api.buck.id}"
  parent_id   = "${aws_api_gateway_rest_api.buck.root_resource_id}"
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
  timeout_milliseconds = 29000

  credentials = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/apigateway_exe"

  integration_http_method = "POST"

  type = "AWS_PROXY"
  uri  = "${aws_lambda_function.buck.invoke_arn}"
}


resource "aws_api_gateway_method" "proxy_root" {
  rest_api_id   = "${aws_api_gateway_rest_api.buck.id}"
  resource_id   = "${aws_api_gateway_rest_api.buck.root_resource_id}"
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_root" {
  rest_api_id = "${aws_api_gateway_rest_api.buck.id}"
  resource_id = "${aws_api_gateway_method.proxy_root.resource_id}"
  http_method = "${aws_api_gateway_method.proxy_root.http_method}"

  credentials = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/apigateway_exe"

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${aws_lambda_function.buck.invoke_arn}"
}

data "aws_caller_identity" "current" {}


resource "aws_lambda_permission" "apigw_buck" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.buck.arn}"
  principal     = "apigateway.amazonaws.com"

  # "arn:aws:execute-api:region:account-id:api-id/stage/METHOD_HTTP_VERB/Resource-path"
  #https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-control-access-using-iam-policies-to-invoke-api.html#

  source_arn = "${aws_api_gateway_rest_api.buck.execution_arn}/*/*/*"
}

resource "aws_api_gateway_deployment" "buck" {
  depends_on = [
    "aws_api_gateway_integration.lambda",
    "aws_api_gateway_integration.lambda_root",
    "aws_lambda_permission.apigw_buck",
  ]

  rest_api_id = "${aws_api_gateway_rest_api.buck.id}"
  stage_name  = "buck"


  # https://medium.com/coryodaniel/til-forcing-terraform-to-deploy-a-aws-api-gateway-deployment-ed36a9f60c1a
  # https://github.com/terraform-providers/terraform-provider-aws/issues/162
  variables {
    deployed_at = "${timestamp()}"
  }

  lifecycle {
    create_before_destroy = true
  }

}




