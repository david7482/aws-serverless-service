provider "aws" {
  region  = var.region
  profile = var.aws_profile
}

# AWS Region for Cloudfront (ACM certs only supports us-east-1)
provider "aws" {
  alias   = "us-east-1"
  region  = "us-east-1"
  profile = var.aws_profile
}