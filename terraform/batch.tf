resource "aws_batch_compute_environment" "jobs_compute" {
  compute_environment_name = "jobs_compute"

  compute_resources {
    max_vcpus = 128

    security_group_ids = [data.aws_security_group.default.id]

    subnets = [
      data.aws_subnets.example.ids[0],
    ]

    type = "FARGATE_SPOT"
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
  platform_capabilities = [
    "FARGATE",
  ]

  container_properties = jsonencode({
    image = aws_ecrpublic_repository.bitcoin-coverage-coverage-worker.repository_uri
    fargatePlatformConfiguration = {
      platformVersion = "LATEST"
    }

    resourceRequirements = [
      {
        type  = "VCPU"
        value = "8",
      },
      {
        type  = "MEMORY"
        value = "16384"
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
      }
    ]

    networkConfiguration = {
      "assignPublicIp" = "ENABLED"
    }

    executionRoleArn = aws_iam_role.ecs_task_execution_role.arn
    jobRoleArn       = aws_iam_role.job_role.arn
  })
  timeout {
    attempt_duration_seconds = 3600
  }
}

resource "aws_batch_job_definition" "mutation_job" {
  name = "mutation-job"
  type = "container"
  platform_capabilities = [
    "FARGATE",
  ]

  container_properties = jsonencode({
    image = aws_ecrpublic_repository.bitcoin-coverage-mutation-worker.repository_uri
    fargatePlatformConfiguration = {
      platformVersion = "LATEST"
    }

    resourceRequirements = [
      {
        type  = "VCPU"
        value = "8",
      },
      {
        type  = "MEMORY"
        value = "16384"
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
      }
    ]

    networkConfiguration = {
      "assignPublicIp" = "ENABLED"
    }

    executionRoleArn = aws_iam_role.ecs_task_execution_role.arn
    jobRoleArn       = aws_iam_role.job_role.arn
  })
  timeout {
    attempt_duration_seconds = 1800
  }
}
