resource "aws_batch_compute_environment" "jobs_compute" {
  compute_environment_name_prefix = "jobs_compute"

  compute_resources {
    max_vcpus          = 128
    security_group_ids = [data.aws_security_group.default.id]

    subnets = [
      data.aws_subnets.example.ids[0],
      data.aws_subnets.example.ids[1],
      data.aws_subnets.example.ids[2],
    ]

    type                = "SPOT"
    instance_role       = aws_iam_instance_profile.ecs_instance_role.arn
    instance_type       = ["c6i.2xlarge"]
    allocation_strategy = "SPOT_PRICE_CAPACITY_OPTIMIZED"
  }
  lifecycle {
    create_before_destroy = true
  }

  type = "MANAGED"
}


resource "aws_batch_job_queue" "coverage_queue" {
  name     = "coverage-queue"
  state    = "ENABLED"
  priority = 1
  compute_environments = [
    aws_batch_compute_environment.jobs_compute.arn
  ]
}

resource "aws_batch_job_definition" "coverage_job" {
  name = "coverage-job"
  type = "container"


  container_properties = jsonencode({
    image      = aws_ecrpublic_repository.bitcoin-coverage-coverage-worker.repository_uri
    privileged = true

    mountPoints = [
      {
        containerPath = "/lib/modules"
        readOnly      = false
        sourceVolume  = "modules"
      }
    ]

    volumes = [
      {
        name = "modules"
        host = {
          sourcePath = "/lib/modules"
        }
      }
    ]

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
        value = aws_s3_bucket.bitcoin-coverage-ccache.id
      },
      {
        name  = "SCCACHE_REGION",
        value = "eu-west-3"
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
      }
    ]

    executionRoleArn = aws_iam_role.job_role.arn
    jobRoleArn       = aws_iam_role.job_role.arn
  })
  timeout {
    attempt_duration_seconds = 5400
  }
}
