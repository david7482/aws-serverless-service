terraform {
  backend "s3" {
    profile              = "david74"
    region               = "us-west-2"
    bucket               = "david74-terraform-remote-state-storage"
    key                  = "terraform.tfstate"
    encrypt              = true
    workspace_key_prefix = "terraform-aws-serverless-service"
  }
}
