resource "aws_batch_compute_environment" "bench" {
  provider                        = aws.compute_region
  compute_environment_name_prefix = "bench-${terraform.workspace}-"

  compute_resources {
    max_vcpus          = 256
    security_group_ids = [data.aws_security_group.compute_security_group.id]

    subnets = [
      data.aws_subnets.batch_subnets.ids[0],
      data.aws_subnets.batch_subnets.ids[1],
      data.aws_subnets.batch_subnets.ids[2],
    ]

    type                = "SPOT"
    instance_role       = aws_iam_instance_profile.ecs_instance_role.arn
    instance_type       = ["c7g.medium"]
    allocation_strategy = "SPOT_PRICE_CAPACITY_OPTIMIZED"

    ec2_configuration {
      image_type        = "ECS_AL2"
      image_id_override = ""
    }
  }
  lifecycle {
    create_before_destroy = true
  }

  type = "MANAGED"
}

resource "aws_batch_job_queue" "bench_queue" {
  name     = "bench-queue-${terraform.workspace}"
  provider = aws.compute_region
  state    = "ENABLED"
  priority = 1
  compute_environments = [
    aws_batch_compute_environment.bench.arn,
  ]
}

resource "aws_batch_job_definition" "bench_job" {
  name     = "bench-job-${terraform.workspace}"
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
    image      = aws_ecrpublic_repository.corecheck-bench-worker.repository_uri
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
        value = "1",
      },
      {
        type  = "MEMORY"
        value = "1500",
      },
    ]

    environment = [
      {
        name  = "AWS_ACCESS_KEY_ID",
        value = var.aws_access_key_id
      },
      {
        name  = "AWS_SECRET_ACCESS_KEY",
        value = var.aws_secret_access_key
      },
      {
        name  = "S3_BUCKET_DATA",
        value = var.corecheck_data_bucket
      },
      {
        name  = "S3_BUCKET_ARTIFACTS",
        value = aws_s3_bucket.corecheck-artifacts.id
      }
    ]

    command = [
      "/entrypoint.sh",
      "Ref::commit",
      "Ref::pr_number",
      "Ref::is_master",
      "Ref::base_commit",
    ]

    executionRoleArn = aws_iam_role.job_role.arn
    jobRoleArn       = aws_iam_role.job_role.arn
  })
  timeout {
    attempt_duration_seconds = 5400
  }
}
