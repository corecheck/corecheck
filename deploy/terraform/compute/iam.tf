data "aws_iam_policy_document" "ec2_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "ecs_instance_role" {
  name               = "ecs_instance_role-${terraform.workspace}"
  assume_role_policy = data.aws_iam_policy_document.ec2_assume_role.json
}

resource "aws_iam_role_policy_attachment" "ecs_instance_role" {
  role       = aws_iam_role.ecs_instance_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_role" {
  name = "ecs_instance_role-${terraform.workspace}"
  role = aws_iam_role.ecs_instance_role.name
}

data "aws_iam_policy_document" "batch_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["batch.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "ecs_task_execution_role" {
  name               = "task_execution_role-${terraform.workspace}"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "job_role_assume_role" {
  statement {
    effect = "Allow"

    # ec2
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "job_role" {
  name               = "job-role-${terraform.workspace}"
  assume_role_policy = data.aws_iam_policy_document.job_role_assume_role.json
}
# allow cloudwatch
resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_cloudwatch" {
  role       = aws_iam_role.job_role.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchFullAccessV2"
}

data "aws_iam_policy_document" "allow_job_timestream_write" {
  count = var.telemetry_backend == "timestream" ? 1 : 0

  statement {
    effect    = "Allow"
    actions   = ["timestream:DescribeEndpoints"]
    resources = ["*"]
  }

  statement {
    effect  = "Allow"
    actions = ["timestream:WriteRecords"]
    resources = [
      "arn:aws:timestream:${var.telemetry_timestream_region}:${data.aws_caller_identity.current.account_id}:database/${var.telemetry_timestream_database}/table/${var.telemetry_timestream_table}",
    ]
  }
}

resource "aws_iam_policy" "job_timestream_write_policy" {
  count       = var.telemetry_backend == "timestream" ? 1 : 0
  name        = "AllowJobTimestreamWritePolicy-${terraform.workspace}"
  description = "Policy for batch jobs to write telemetry metrics to timestream"
  policy      = data.aws_iam_policy_document.allow_job_timestream_write[0].json
}

resource "aws_iam_role_policy_attachment" "job_timestream_write_policy_attachment" {
  count      = var.telemetry_backend == "timestream" ? 1 : 0
  role       = aws_iam_role.job_role.name
  policy_arn = aws_iam_policy.job_timestream_write_policy[0].arn
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}
