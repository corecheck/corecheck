data "aws_vpc" "batch_vpc" {
  provider = aws.compute_region
  default  = true
}

data "aws_subnets" "batch_subnets" {
  provider = aws.compute_region
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.batch_vpc.id]
  }
}

resource "aws_s3_bucket" "corecheck-ccache" {
  provider = aws.compute_region
  bucket   = "corecheck-ccache-${terraform.workspace}"
  force_destroy = true
}


# remove objects after 30days
resource "aws_s3_bucket_lifecycle_configuration" "corecheck-ccache" {
  provider = aws.compute_region
  bucket   = aws_s3_bucket.corecheck-ccache.id

  rule {
    id     = "corecheck-ccache"
    status = "Enabled"
    expiration {
      days = 30
    }
  }
}
resource "aws_s3_bucket" "corecheck-artifacts" {
  provider = aws.compute_region
  bucket   = "corecheck-artifacts-${terraform.workspace}"
  force_destroy = true
}

resource "aws_s3_bucket_lifecycle_configuration" "corecheck-artifacts" {
  provider = aws.compute_region
  bucket   = aws_s3_bucket.corecheck-artifacts.id

  rule {
    id     = "corecheck-artifacts"
    status = "Enabled"
    expiration {
      days = 3
    }
  }
}

data "aws_security_group" "compute_security_group" {
  provider = aws.compute_region
  name     = "default"
}
