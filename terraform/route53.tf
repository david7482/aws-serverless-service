################################
# Locals
################################
locals {
  route53_record_env = var.env == "production" ? "" : "-${var.env}"
}

################################
# Route53
################################
data "aws_route53_zone" "main" {
  name = var.domain
}

locals {
  route53_record_alias_name    = aws_alb.alb.dns_name
  route53_record_alias_zone_id = aws_alb.alb.zone_id
}

resource "aws_route53_record" "service" {
  zone_id = data.aws_route53_zone.main.zone_id
  name    = format(
    "%s%s.%s",
    var.app_name,
    local.route53_record_env,
    data.aws_route53_zone.main.name,
  )
  type = "A"

  alias {
    evaluate_target_health = true
    name                   = local.route53_record_alias_name
    zone_id                = local.route53_record_alias_zone_id
  }
}

resource "aws_route53_record" "slide" {
  zone_id = data.aws_route53_zone.main.zone_id
  name    = local.cloudfront_fqdn
  type    = "A"

  alias {
    evaluate_target_health = false
    name                   = aws_cloudfront_distribution.slide.domain_name
    zone_id                = aws_cloudfront_distribution.slide.hosted_zone_id
  }
}