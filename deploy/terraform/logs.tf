resource "aws_cloudwatch_log_group" "batch_logs" {
  name = "/aws/batch/job-${terraform.workspace}"

  retention_in_days = 7

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = false
  }
}