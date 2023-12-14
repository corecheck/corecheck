terraform {
  required_version = ">= 0.14.0"
  required_providers {
    aws = {
      source                = "hashicorp/aws"
      version               = "5.15.0"
    }
  }
}

provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}
