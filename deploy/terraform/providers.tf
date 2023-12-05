terraform {
  required_version = ">= 0.14.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.15.0"
    }
  }

  backend "s3" {
    bucket = "bitcoin-coverage-state"
    key    = "terraform.tfstate"
    region = "eu-west-3"
  }
}


provider "aws" {
  region = "eu-west-3"
}

provider "aws" {
  alias  = "compute_region"
  region = "ap-south-1"
}

data "aws_region" "compute_region" {
  provider = aws.compute_region
}
provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}