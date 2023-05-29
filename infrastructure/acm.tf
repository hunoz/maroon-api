resource "aws_acm_certificate" "api_cert" {
  domain_name       = "${local.dns_name}.maroon.gtech.dev"
  validation_method = "DNS"
}

data "aws_route53_zone" "maroon_hosted_zone" {
  name         = "maroon.gtech.dev"
  private_zone = false
}

resource "aws_route53_record" "dns_validation_records" {
  for_each = {
    for dvo in aws_acm_certificate.api_cert.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.maroon_hosted_zone.zone_id
}

resource "aws_acm_certificate_validation" "api_cert_validation" {
  certificate_arn         = aws_acm_certificate.api_cert.arn
  validation_record_fqdns = [for record in aws_route53_record.dns_validation_records : record.fqdn]
}

resource "aws_route53_record" "api_dns_record" {
  allow_overwrite = true
  name            = aws_apigatewayv2_domain_name.maroon_api_domain_name.domain_name
  type            = "A"
  zone_id         = data.aws_route53_zone.maroon_hosted_zone.zone_id

  alias {
    name                   = aws_apigatewayv2_domain_name.maroon_api_domain_name.domain_name_configuration[0].target_domain_name
    zone_id                = aws_apigatewayv2_domain_name.maroon_api_domain_name.domain_name_configuration[0].hosted_zone_id
    evaluate_target_health = false
  }
}