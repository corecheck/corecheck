provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}

resource "aws_ecrpublic_repository" "bitcoin-coverage-mutation-worker" {
  provider        = aws.us_east_1
  repository_name = "bitcoin-coverage-mutation-worker"
}

resource "aws_ecrpublic_repository" "bitcoin-coverage-coverage-worker" {
  provider        = aws.us_east_1
  repository_name = "bitcoin-coverage-coverage-worker"
}
