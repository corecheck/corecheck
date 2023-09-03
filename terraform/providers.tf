terraform {
  required_version = ">= 0.14.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.15.0"
    }
  }

  backend "s3" {
    bucket = "bitcoin-coverage-tfstate"
    key    = "terraform.tfstate"
    region = "eu-west-1"
  }
}


provider "aws" {
  region = "eu-west-1"
}