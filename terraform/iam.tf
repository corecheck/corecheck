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
  name               = "ecs_instance_role"
  assume_role_policy = data.aws_iam_policy_document.ec2_assume_role.json
}

resource "aws_iam_role_policy_attachment" "ecs_instance_role" {
  role       = aws_iam_role.ecs_instance_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_role" {
  name = "ecs_instance_role"
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

data "aws_iam_role" "aws_service_role_batch" {
  name = "AWSServiceRoleForBatch"
}

resource "aws_iam_role" "ecs_task_execution_role" {
  name               = "task_execution_role"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "job_role_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "job_role" {
  name               = "job-role"
  assume_role_policy = data.aws_iam_policy_document.job_role_assume_role.json
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

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

data "aws_iam_policy_document" "job-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.bitcoin-coverage-data.arn,
      "${aws_s3_bucket.bitcoin-coverage-data.arn}/*",
      aws_s3_bucket.bitcoin-coverage-cache.arn,
      "${aws_s3_bucket.bitcoin-coverage-cache.arn}/*",
      aws_s3_bucket.bitcoin-coverage-ccache.arn,
      "${aws_s3_bucket.bitcoin-coverage-ccache.arn}/*",
    ]
  }
}

resource "aws_iam_policy" "job-role-policy" {
  name        = "job-role-policy"
  description = "Allows ECS tasks to call AWS services on your behalf."
  policy      = data.aws_iam_policy_document.job-role-policy.json
}

resource "aws_iam_role_policy_attachment" "job-role-policy" {
  role       = aws_iam_role.job_role.name
  policy_arn = aws_iam_policy.job-role-policy.arn
}
