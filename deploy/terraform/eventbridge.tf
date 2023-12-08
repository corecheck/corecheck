# event bridge bus default
data "aws_cloudwatch_event_bus" "default" {
  name     = "default"
  provider = aws.compute_region
}

# example
# {
#   "version": "0",
#   "id": "38222ec2-34ce-02ac-86ed-9cd3d65cb5be",
#   "detail-type": "start-jobs",
#   "source": "corecheck",
#   "account": "338220712891",
#   "time": "2023-12-08T22:47:05Z",
#   "region": "ap-south-1",
#   "resources": [],
#   "detail": {
#     "commit": "123",
#     "is_master": false,
#     "pr_num": 1234
#   }
# }

resource "aws_cloudwatch_event_rule" "start_jobs" {
  name           = "start-jobs"
  description    = "start jobs"
  event_bus_name = data.aws_cloudwatch_event_bus.default.name
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


# create event bridge target
resource "aws_cloudwatch_event_target" "coverage_target" {
  rule = aws_cloudwatch_event_rule.start_jobs.name
  arn  = aws_batch_job_queue.coverage_queue.arn

  batch_target {
    job_definition = aws_batch_job_definition.coverage_job.arn
    job_name       = "coverage"
  }

  input = <<INPUT
{
  "commit": "$${detail.commit}",
  "is_master": "$${detail.is_master}",
  "pr_num": "$${detail.pr_num}"
}
INPUT
}

# another batch target
resource "aws_cloudwatch_event_target" "sonar_target" {
  rule = aws_cloudwatch_event_rule.start_jobs.name
  arn  = aws_batch_job_queue.sonar_queue.arn

  batch_target {
    job_definition = aws_batch_job_definition.sonar_job.arn
    job_name       = "sonar"
  }

  input = <<INPUT
{
  "commit": "$${detail.commit}",
  "is_master": "$${detail.is_master}",
  "pr_num": "$${detail.pr_num}"
}
INPUT
}
