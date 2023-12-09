# event bridge bus default
data "aws_cloudwatch_event_bus" "default" {
  name     = "default"
  provider = aws.compute_region
}

resource "aws_cloudwatch_event_rule" "start_jobs" {
  name           = "start-jobs"
  description    = "start jobs"
  event_bus_name = data.aws_cloudwatch_event_bus.default.name
  provider       = aws.compute_region
  event_pattern  = <<PATTERN
{
  "source": [
    "corecheck"
  ],
  "detail-type": [
    "start-jobs"
  ]
}

PATTERN
}


data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "event_bridge_role" {
  name               = "event_bridge_role-${terraform.workspace}"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}


resource "aws_iam_role_policy_attachment" "event_bridge_policy_attachment" {
  role       = aws_iam_role.event_bridge_role.name
  policy_arn = "arn:aws:iam::aws:policy/AWSBatchFullAccess"
}

# create event bridge target
resource "aws_cloudwatch_event_target" "coverage_target" {
  rule     = aws_cloudwatch_event_rule.start_jobs.name
  arn      = aws_batch_job_queue.coverage_queue.arn
  role_arn = aws_iam_role.event_bridge_role.arn
  provider = aws.compute_region

  batch_target {
    job_definition = aws_batch_job_definition.coverage_job.arn
    job_name       = "coverage"
  }

  input_path = "$.detail"
}


# another batch target
resource "aws_cloudwatch_event_target" "sonar_target" {
  rule     = aws_cloudwatch_event_rule.start_jobs.name
  arn      = aws_batch_job_queue.sonar_queue.arn
  role_arn = aws_iam_role.event_bridge_role.arn
  provider = aws.compute_region

  batch_target {
    job_definition = aws_batch_job_definition.sonar_job.arn
    job_name       = "sonar"
  }

  input_path = "$.detail"
}


# logging event bridge
resource "aws_cloudwatch_log_group" "event_bridge" {
  name              = "/aws/events/event_bridge"
  retention_in_days = 7
  provider          = aws.compute_region
}
