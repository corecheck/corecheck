resource "aws_batch_compute_environment" "sonar_compute" {
  provider                        = aws.compute_region
  compute_environment_name_prefix = "sonar-"

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
    instance_type       = ["c6a.2xlarge"]
    allocation_strategy = "SPOT_PRICE_CAPACITY_OPTIMIZED"
  }
  lifecycle {
    create_before_destroy = true
  }

  type = "MANAGED"
}


resource "aws_batch_job_queue" "sonar_queue" {
  name     = "sonar-queue"
  provider = aws.compute_region
  state    = "ENABLED"
  priority = 1
  compute_environments = [
    aws_batch_compute_environment.sonar_compute.arn
  ]
}
resource "aws_batch_job_definition" "sonar_job" {
  name     = "sonar-job"
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
    image      = aws_ecrpublic_repository.corecheck-sonar-worker.repository_uri

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
