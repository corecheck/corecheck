terraform {
  required_version = ">= 0.14.0"
  required_providers {
    aws = {
      configuration_aliases = [aws.compute_region]
      source                = "hashicorp/aws"
      version               = "5.90.0"
    }
  }
}
