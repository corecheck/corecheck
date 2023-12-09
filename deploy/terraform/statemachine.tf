
# state machine role
resource "aws_iam_role" "state_machine_role" {
  name = "state_machine_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "states.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

# state machine policy (batch + lambda), create managed-rule
#   - events:PutTargets
#   - events:PutRule
#   - events:DescribeRule
resource "aws_iam_role_policy" "state_machine_policy" {
  name = "state_machine_policy"
  role = aws_iam_role.state_machine_role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "lambda:InvokeFunction"
      ],
      "Effect": "Allow",
      "Resource": "*"
    },
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Effect": "Allow",
      "Resource": "*"
    },
    {
      "Action": [
        "batch:SubmitJob",
        "batch:TerminateJob",
        "batch:DescribeJobs",
        "batch:DescribeJobDefinitions",
        "batch:DescribeJobQueues",
        "batch:RegisterJobDefinition"
      ],
      "Effect": "Allow",
      "Resource": "*"
    },
    {
      "Action": [
        "events:PutTargets",
        "events:PutRule",
        "events:DescribeRule"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

# statemachine resource
resource "aws_sfn_state_machine" "state_machine" {
  name     = "start-jobs"
  role_arn = aws_iam_role.state_machine_role.arn
  provider = aws.compute_region

  definition = <<EOF
{
  "Comment": "A description of my state machine",
  "StartAt": "Start coverage",
  "States": {
    "Start coverage": {
      "Type": "Task",
      "Resource": "arn:aws:states:::batch:submitJob.sync",
      "Parameters": {
        "Parameters.$": "$",
        "JobDefinition": "${aws_batch_job_definition.coverage_job.arn}",
        "JobName": "coverage",
        "JobQueue": "${aws_batch_job_queue.coverage_queue.arn}"
      },
      "End": true
    }
  }
}
EOF
}
