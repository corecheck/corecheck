resource "aws_cloudwatch_log_group" "batch_logs" {
  name = "/aws/batch/job-${terraform.workspace}"

  retention_in_days = 7

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = false
  }
}

resource "aws_cloudwatch_log_group" "test_results" {
  provider          = aws.compute_region
  name              = "/corecheck/test-results/${terraform.workspace}"
  retention_in_days = 1096

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = false
  }
}