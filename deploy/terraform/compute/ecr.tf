resource "aws_ecrpublic_repository" "corecheck-coverage-worker" {
  provider        = aws.us_east_1
  repository_name = "corecheck-coverage-worker-${terraform.workspace}"
  force_destroy = true
}

resource "aws_ecrpublic_repository" "corecheck-bench-worker" {
  provider        = aws.us_east_1
  repository_name = "corecheck-bench-worker-${terraform.workspace}"
  force_destroy = true
}

resource "aws_ecrpublic_repository" "corecheck-sonar-worker" {
  provider        = aws.us_east_1
  repository_name = "corecheck-sonar-worker-${terraform.workspace}"
  force_destroy = true
}

resource "aws_ecrpublic_repository" "corecheck-mutation-worker" {
  provider        = aws.us_east_1
  repository_name = "corecheck-mutation-worker-${terraform.workspace}"
  force_destroy = true
}
