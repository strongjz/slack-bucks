resource "aws_lambda_function" "buck" {
  function_name = "Slackbuck"

  s3_bucket = "terraform-serverless-buck"
  s3_key = "${var.app_version}/buck.zip"

  handler = "main"
  runtime = "go1.x"

  timeout = "900"

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

# See also the following AWS managed policy: AWSLambdaBasicExecutionRole
resource "aws_iam_policy" "lambda_logging" {
  name = "lambda_logging"
  path = "/"
  description = "IAM policy for logging from a lambda"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role = "${aws_iam_role.lambda_exec.name}"
  policy_arn = "${aws_iam_policy.lambda_logging.arn}"
}

resource "aws_api_gateway_account" "buck" {
  cloudwatch_role_arn = "${aws_iam_role.cloudwatch.arn}"
}

resource "aws_iam_role" "cloudwatch" {
  name = "api_gateway_cloudwatch_global"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "cloudwatch" {
  name = "default"
  role = "${aws_iam_role.cloudwatch.id}"

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:DescribeLogGroups",
                "logs:DescribeLogStreams",
                "logs:PutLogEvents",
                "logs:GetLogEvents",
                "logs:FilterLogEvents"
            ],
            "Resource": "*"
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
  path_part = "buck"
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


resource "aws_route53_delegation_set" "bucks" {
  reference_name = "bucks"
}

resource "aws_route53_zone" "bucks" {
  name              = "${var.domain}"
  delegation_set_id = "${aws_route53_delegation_set.bucks.id}"
}

resource "aws_acm_certificate" "bucks" {
  provider = "aws.east"
  domain_name       = "${var.domain}"
  validation_method = "DNS"


  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_route53_record" "cert_validation" {
  name    = "${aws_acm_certificate.bucks.domain_validation_options.0.resource_record_name}"
  type    = "${aws_acm_certificate.bucks.domain_validation_options.0.resource_record_type}"
  zone_id = "${aws_route53_zone.bucks.id}"
  records = ["${aws_acm_certificate.bucks.domain_validation_options.0.resource_record_value}"]
  ttl     = 60
}

resource "aws_acm_certificate_validation" "bucks" {
  provider = "aws.east"
  certificate_arn         = "${aws_acm_certificate.bucks.arn}"
  validation_record_fqdns = ["${aws_route53_record.cert_validation.fqdn}"]
}


resource "aws_api_gateway_domain_name" "bucks" {
  certificate_arn = "${aws_acm_certificate_validation.bucks.certificate_arn}"
  domain_name     = "${var.domain}"
}


# Example DNS record using Route53.
# Route53 is not specifically required; any DNS host can be used.
resource "aws_route53_record" "bucks" {
  name    = "${aws_api_gateway_domain_name.bucks.domain_name}"
  type    = "A"
  zone_id = "${aws_route53_zone.bucks.id}"

  alias {
    evaluate_target_health = true
    name                   = "${aws_api_gateway_domain_name.bucks.cloudfront_domain_name}"
    zone_id                = "${aws_api_gateway_domain_name.bucks.cloudfront_zone_id}"
  }
}


output "base_url" {
  value = "${aws_api_gateway_deployment.buck.invoke_url}"
}

