# sqs queue
resource "aws_sqs_queue" "corecheck_queue" {
  name = "corecheck-queue-${terraform.workspace}"
}