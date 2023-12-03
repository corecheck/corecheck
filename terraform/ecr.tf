resource "aws_ecrpublic_repository" "corecheck-coverage-worker" {
  provider        = aws.us_east_1
  repository_name = "corecheck-coverage-worker"
}

resource "aws_ecrpublic_repository" "corecheck-bench-worker" {
  provider        = aws.us_east_1
  repository_name = "corecheck-bench-worker"
}

resource "aws_ecrpublic_repository" "corecheck-sonar-worker" {
  provider        = aws.us_east_1
  repository_name = "corecheck-sonar-worker"
}

