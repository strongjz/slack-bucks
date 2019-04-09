provider "aws" {
  region = "us-east-1"
}


variable "app_version" {}

variable "verificationToken" {}

variable "oauthToken" {}



resource "aws_lambda_function" "buck" {
  function_name = "Slackbuck"

  s3_bucket = "terraform-serverless-buck"
  s3_key = "v${var.app_version}/buck.zip"

  handler = "main"
  runtime = "go1.x"

  role = "${aws_iam_role.lambda_exec.arn}"

  environment {
    variables = {
      verificationToken = "${var.verificationToken}"
      oauthToken = "${var.oauthToken}"
    }
  }


}

# IAM role which dictates what other AWS services the Lambda function
# may access.
resource "aws_iam_role" "lambda_exec" {
  name = "serverless_buck_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_api_gateway_rest_api" "buck" {
  name = "Serverlessbuck"
  description = "Terraform Serverless Application buck"
}

resource "aws_api_gateway_resource" "proxy" {
  rest_api_id = "${aws_api_gateway_rest_api.buck.id}"
  parent_id = "${aws_api_gateway_rest_api.buck.root_resource_id}"
  path_part = "{proxy+}"
}

resource "aws_api_gateway_method" "proxy" {
  rest_api_id = "${aws_api_gateway_rest_api.buck.id}"
  resource_id = "${aws_api_gateway_resource.proxy.id}"
  http_method = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda" {
  rest_api_id = "${aws_api_gateway_rest_api.buck.id}"
  resource_id = "${aws_api_gateway_method.proxy.resource_id}"
  http_method = "${aws_api_gateway_method.proxy.http_method}"

  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "${aws_lambda_function.buck.invoke_arn}"
}

resource "aws_api_gateway_method" "proxy_root" {
  rest_api_id = "${aws_api_gateway_rest_api.buck.id}"
  resource_id = "${aws_api_gateway_rest_api.buck.root_resource_id}"
  http_method = "ANY"
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


output "base_url" {
  value = "${aws_api_gateway_deployment.buck.invoke_url}"
}

