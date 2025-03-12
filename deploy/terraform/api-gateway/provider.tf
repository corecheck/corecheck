terraform {
  required_version = ">= 0.14.0"
  required_providers {
    aws = {
      configuration_aliases = [aws.us_east_1]
      source                = "hashicorp/aws"
      version               = "5.15.0"
    }
  }
}
