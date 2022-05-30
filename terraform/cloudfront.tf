locals {
  cloudfront_price_class = "PriceClass_200"
  cloudfront_fqdn        = "slide.david74.dev"
  cloudfront_comment     = "${var.app_name}-${var.env}"
}

data "aws_acm_certificate" "cloudfront_cert" {
  provider = aws.us-east-1
  domain   = "*.david74.dev"
}

resource "aws_cloudfront_origin_access_identity" "origin_access_identity" {
  comment = local.cloudfront_comment
}

resource "aws_cloudfront_distribution" "slide" {
  enabled             = true
  is_ipv6_enabled     = true
  comment             = local.cloudfront_comment
  price_class         = local.cloudfront_price_class
  wait_for_deployment = false
  aliases             = [local.cloudfront_fqdn]

  viewer_certificate {
    acm_certificate_arn      = data.aws_acm_certificate.cloudfront_cert.arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.1_2016"
  }

  origin {
    domain_name = aws_s3_bucket.main.bucket_regional_domain_name
    origin_id   = aws_s3_bucket.main.id

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.origin_access_identity.cloudfront_access_identity_path
    }
  }

  default_cache_behavior {
    target_origin_id       = aws_s3_bucket.main.id
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    compress               = true
    viewer_protocol_policy = "redirect-to-https"
    default_ttl            = 3600
    min_ttl                = 0
    max_ttl                = 3600

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  tags = {
    Name        = "${var.app_name}-${var.env}"
    Environment = var.env
  }
}
