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

