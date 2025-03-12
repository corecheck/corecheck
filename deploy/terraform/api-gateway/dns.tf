data "aws_route53_zone" "zone" {
  name = var.dns_name
}

locals {
  api_name = terraform.workspace == "default" ? "api" : "api-${terraform.workspace}"
  api_domain = "${local.api_name}.${var.dns_name}"

  datadog_proxy_name = terraform.workspace == "default" ? "datadog-proxy" : "datadog-proxy-${terraform.workspace}"
  datadog_proxy_domain = "${local.datadog_proxy_name}.${var.dns_name}"
}

resource "aws_acm_certificate" "api_gw" {
  domain_name = local.api_domain
  validation_method = "DNS"
  provider = aws.us_east_1
}

# route 53 record for api gateway certificate validation
resource "aws_route53_record" "api_gw_validation" {
  for_each = {
      for dvo in aws_acm_certificate.api_gw.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  zone_id = data.aws_route53_zone.zone.zone_id
  name    = each.value.name
  type    = each.value.type
  records = [each.value.record]
  ttl = 60

  depends_on = [
    aws_acm_certificate.api_gw,
  ]
}

resource "aws_acm_certificate_validation" "api_gw" {
  provider = aws.us_east_1
  certificate_arn         = aws_acm_certificate.api_gw.arn
  validation_record_fqdns = [for record in aws_route53_record.api_gw_validation : record.fqdn]

    depends_on = [
        aws_route53_record.api_gw_validation,
    ]
}

resource "aws_api_gateway_domain_name" "api_gw" {
  domain_name = local.api_domain
  certificate_arn = "${aws_acm_certificate.api_gw.arn}"
  certificate_chain = "${aws_acm_certificate.api_gw.certificate_chain}"
}

# custom domain name mapping
resource "aws_api_gateway_base_path_mapping" "api_gw" {
  domain_name = local.api_domain
  api_id = aws_api_gateway_rest_api.api.id
  stage_name = aws_api_gateway_deployment.api.stage_name

  depends_on = [
    aws_api_gateway_domain_name.api_gw,
  ]
}

resource "aws_route53_record" "api_gw" {
  zone_id = data.aws_route53_zone.zone.zone_id
  name    = local.api_domain
  type    = "A"
  alias {
    name                   = "${aws_api_gateway_domain_name.api_gw.cloudfront_domain_name}"
    zone_id                = "${aws_api_gateway_domain_name.api_gw.cloudfront_zone_id}"
    evaluate_target_health = false
  }
  depends_on = [
    aws_api_gateway_domain_name.api_gw,
  ]
}


resource "aws_acm_certificate" "datadog_proxy" {
  domain_name = local.datadog_proxy_domain
  validation_method = "DNS"
  provider = aws.us_east_1
}

# route 53 record for datadog proxy certificate validation
resource "aws_route53_record" "datadog_proxy_validation" {
  for_each = {
      for dvo in aws_acm_certificate.datadog_proxy.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  zone_id = data.aws_route53_zone.zone.zone_id
  name    = each.value.name
  type    = each.value.type
  records = [each.value.record]
  ttl = 60

  depends_on = [
    aws_acm_certificate.datadog_proxy,
  ]
}

resource "aws_acm_certificate_validation" "datadog_proxy" {
  provider = aws.us_east_1
  certificate_arn         = aws_acm_certificate.datadog_proxy.arn
  validation_record_fqdns = [for record in aws_route53_record.datadog_proxy_validation : record.fqdn]

    depends_on = [
        aws_route53_record.datadog_proxy_validation,
    ]
}

resource "aws_api_gateway_domain_name" "datadog_proxy" {
  domain_name = local.datadog_proxy_domain
  certificate_arn = "${aws_acm_certificate.datadog_proxy.arn}"
  certificate_chain = "${aws_acm_certificate.datadog_proxy.certificate_chain}"
}

# custom domain name mapping
resource "aws_api_gateway_base_path_mapping" "datadog_proxy" {
  domain_name = local.datadog_proxy_domain
  api_id = aws_api_gateway_rest_api.datadog_proxy.id
  stage_name = aws_api_gateway_deployment.datadog_proxy.stage_name

  depends_on = [
    aws_api_gateway_domain_name.datadog_proxy,
  ]
}

resource "aws_route53_record" "datadog_proxy" {
  zone_id = data.aws_route53_zone.zone.zone_id
  name    = local.datadog_proxy_domain
  type    = "A"
  alias {
    name                   = "${aws_api_gateway_domain_name.datadog_proxy.cloudfront_domain_name}"
    zone_id                = "${aws_api_gateway_domain_name.datadog_proxy.cloudfront_zone_id}"
    evaluate_target_health = false
  }
  depends_on = [
    aws_api_gateway_domain_name.datadog_proxy,
  ]
}

