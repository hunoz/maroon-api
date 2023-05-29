terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.48.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 3.31.0"
    }
  }
  backend "s3" {
    bucket               = "terraform-maroon-api"
    key                  = "maroon-api"
    region               = "us-west-2"
    workspace_key_prefix = "environments"
  }
  required_version = ">= 1.3.1"
}

locals {
  dns_name = "api"
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}