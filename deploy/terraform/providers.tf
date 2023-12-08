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
  alias  = "default"
  region = "eu-west-3"
}

provider "aws" {
  alias  = "compute_region"
  region = "ap-south-1"
}

data "aws_region" "compute_region" {
  provider = aws.compute_region
}

data "aws_region" "default" {
  provider = aws.default
}

provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}

resource "local_file" "hosts" {
  content  = <<EOF
db ansible_host=${aws_instance.db.public_ip} ansible_ssh_user=ubuntu 
EOF
  filename = "../ansible/hosts.ini"

}

