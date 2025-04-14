provider "aws" {
  region = "eu-west-3"
}

provider "aws" {
  alias  = "compute_region"
  region = "ap-south-1"
}

resource "aws_s3_bucket" "terraform_state" {
  bucket = "bitcoin-coverage-state-${terraform.workspace}"
  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket_versioning" "terraform_state" {
    bucket = aws_s3_bucket.terraform_state.id

    versioning_configuration {
      status = "Enabled"
    }
}

# S3 bucket to store built API Gateway Lambda functions
resource "aws_s3_bucket" "corecheck-lambdas" {
  bucket   = "corecheck-compute-lambdas-${terraform.workspace}"
  provider = aws.compute_region
}

# S3 bucket to store built compute Lambda functions
resource "aws_s3_bucket" "corecheck-lambdas-api" {
  bucket = "corecheck-api-lambdas-${terraform.workspace}"
}
