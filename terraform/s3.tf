resource "aws_s3_bucket" "main" {
  bucket = "david74-${var.app_name}-${var.env}"
  tags   = {
    Name        = "${var.app_name}-${var.env}"
    Environment = var.env
  }
}

resource "aws_s3_bucket_acl" "example_bucket_acl" {
  bucket = aws_s3_bucket.main.id
  acl    = "public-read"
}