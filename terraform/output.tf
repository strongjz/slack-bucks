output "base_url" {
  value = "${aws_api_gateway_deployment.buck.invoke_url}"
}
