resource "aws_s3_bucket" "bitcoin-coverage-data" {
  bucket = "bitcoin-coverage-data-${terraform.workspace}"
}

resource "aws_s3_bucket_ownership_controls" "bitcoin-coverage-data" {
  bucket = aws_s3_bucket.bitcoin-coverage-data.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}
resource "aws_s3_bucket_public_access_block" "bitcoin-coverage-data-public" {
  bucket = aws_s3_bucket.bitcoin-coverage-data.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}
resource "aws_s3_bucket_acl" "bitcoin-coverage-data-public" {
  depends_on = [aws_s3_bucket_public_access_block.bitcoin-coverage-data-public, aws_s3_bucket_ownership_controls.bitcoin-coverage-data]
  bucket     = aws_s3_bucket.bitcoin-coverage-data.id
  acl        = "public-read"
}

resource "aws_s3_bucket_policy" "bitcoin-coverage-data-public" {
  depends_on = [aws_s3_bucket_public_access_block.bitcoin-coverage-data-public, aws_s3_bucket_ownership_controls.bitcoin-coverage-data]
  bucket     = aws_s3_bucket.bitcoin-coverage-data.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "PublicReadGetObject"
        Effect    = "Allow"
        Principal = "*"
        Action = [
          "s3:GetObject"
        ]
        Resource = [
          "${aws_s3_bucket.bitcoin-coverage-data.arn}/*"
        ]
      }
    ]
  })
}

# remove objects after 30days
resource "aws_s3_bucket_lifecycle_configuration" "bitcoin-coverage-data" {
  bucket = aws_s3_bucket.bitcoin-coverage-data.id

  rule {
    id     = "bitcoin-coverage-data"
    status = "Enabled"
    expiration {
      days = 30
    }
  }
}

resource "aws_s3_bucket" "corecheck-lambdas" {
  bucket = "corecheck-lambdas-${terraform.workspace}"
}