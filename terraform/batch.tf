data "aws_vpc" "batch_vpc" {
  provider = aws.compute_region
  default  = true
}

data "aws_subnets" "batch_subnets" {
  provider = aws.compute_region
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.batch_vpc.id]
  }
}

resource "aws_s3_bucket" "corecheck-ccache" {
  provider = aws.compute_region
  bucket   = "corecheck-ccache"
}


# remove objects after 30days
resource "aws_s3_bucket_lifecycle_configuration" "corecheck-ccache" {
  provider = aws.compute_region
  bucket   = aws_s3_bucket.corecheck-ccache.id

  rule {
    id     = "corecheck-ccache"
    status = "Enabled"
    expiration {
      days = 30
    }
  }
}
resource "aws_s3_bucket" "corecheck-artifacts" {
  provider = aws.compute_region
  bucket   = "corecheck-artifacts"
}

resource "aws_s3_bucket_lifecycle_configuration" "corecheck-artifacts" {
  provider = aws.compute_region
  bucket   = aws_s3_bucket.corecheck-artifacts.id

  rule {
    id     = "corecheck-artifacts"
    status = "Enabled"
    expiration {
      days = 3
    }
  }
}

data "aws_security_group" "compute_security_group" {
  provider = aws.compute_region
  name     = "default"
}

resource "aws_batch_compute_environment" "jobs_compute" {
  provider                        = aws.compute_region
  compute_environment_name_prefix = "coverage-"

  compute_resources {
    max_vcpus          = 32
    security_group_ids = [data.aws_security_group.compute_security_group.id]

    subnets = [
      data.aws_subnets.batch_subnets.ids[0],
      data.aws_subnets.batch_subnets.ids[1],
      data.aws_subnets.batch_subnets.ids[2],
    ]

    type                = "SPOT"
    instance_role       = aws_iam_instance_profile.ecs_instance_role.arn
    instance_type       = ["c7g.2xlarge"]
    allocation_strategy = "SPOT_PRICE_CAPACITY_OPTIMIZED"
  }
  lifecycle {
    create_before_destroy = true
  }

  type = "MANAGED"
}


resource "aws_batch_job_queue" "coverage_queue" {
  name     = "coverage-queue"
  provider = aws.compute_region
  state    = "ENABLED"
  priority = 1
  compute_environments = [
    aws_batch_compute_environment.jobs_compute.arn
  ]
}
resource "aws_batch_job_definition" "coverage_job" {
  name     = "coverage-job"
  type     = "container"
  provider = aws.compute_region

  retry_strategy {
    evaluate_on_exit {
      on_status_reason = "Host EC2*"
      action           = "RETRY"
    }

    evaluate_on_exit {
      on_reason = "*"
      action    = "EXIT"
    }

    attempts = 5
  }

  container_properties = jsonencode({
    image      = aws_ecrpublic_repository.corecheck-coverage-worker.repository_uri

    resourceRequirements = [
      {
        type  = "VCPU"
        value = "8",
      },
      {
        type  = "MEMORY"
        value = "15000",
      }
    ]

    environment = [
      {
        name  = "SCCACHE_BUCKET",
        value = aws_s3_bucket.corecheck-ccache.id
      },
      {
        name  = "SCCACHE_REGION",
        value = data.aws_region.compute_region.name
      },
      {
        name  = "AWS_ACCESS_KEY_ID",
        value = var.aws_access_key_id
      },
      {
        name  = "AWS_SECRET_ACCESS_KEY",
        value = var.aws_secret_access_key
      },
      {
        name  = "SONAR_TOKEN",
        value = var.sonar_token
      },
      {
        name  = "S3_BUCKET_DATA",
        value = aws_s3_bucket.bitcoin-coverage-data.id
      },
      {
        name = "S3_BUCKET_ARTIFACTS",
        value = aws_s3_bucket.corecheck-artifacts.id
      }
    ]

    executionRoleArn = aws_iam_role.job_role.arn
    jobRoleArn       = aws_iam_role.job_role.arn
  })
  timeout {
    attempt_duration_seconds = 5400
  }
}
