terraform {
  required_version = ">= 0.14.0"
  required_providers {
    aws = {
      configuration_aliases = [aws.compute_region]
      source                = "hashicorp/aws"
      version               = "5.15.0"
    }
  }
}

provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}

data "aws_region" "compute_region" {
  provider = aws.compute_region
}
