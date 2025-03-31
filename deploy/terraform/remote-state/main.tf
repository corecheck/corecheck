provider "aws" {
  region = "eu-west-3"
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
