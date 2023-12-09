resource "aws_imagebuilder_component" "bitcoin-coverage-component" {
  provider = aws.compute_region
  name     = "CPUAffinity-${terraform.workspace}"
  platform = "Linux"
  version  = "1.0.4"

  lifecycle {
    create_before_destroy = true
  }

  data = yamlencode(
    {
      "schemaVersion" = "1.0"
      "description"   = "CPUAffinity"
      "phases" = [
        {
          "name" = "build"
          "steps" = [
            {
              "name"   = "CPUAffinity"
              "action" = "ExecuteBash"
              "inputs" = {
                "commands" = [
                  "echo 'CPUAffinity=0' >> /etc/systemd/system.conf",
                  "echo 'kernel.randomize_va_space=0' >> /etc/sysctl.conf"
                ]
              }
            }
          ]
        }
      ]
    }
  )
}

# arm 64 Amazon Linux 2
data "aws_ami" "amazon-linux-2" {
  provider    = aws.compute_region
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn2-ami-ecs-hvm*"]
  }

  filter {
    name   = "architecture"
    values = ["arm64"]
  }
}

resource "aws_imagebuilder_image_recipe" "bitcoin-coverage-recipe" {
  provider     = aws.compute_region
  name         = "bitcoin-coverage-recipe-${terraform.workspace}"
  version      = "1.0.10"
  parent_image = data.aws_ami.amazon-linux-2.id
  block_device_mapping {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = false
      volume_size           = 30
      volume_type           = "gp3"
      iops                  = 3000
    }
  }
  component {
    component_arn = aws_imagebuilder_component.bitcoin-coverage-component.arn
  }

  lifecycle {
    create_before_destroy = true
  }
}

# distribution
resource "aws_imagebuilder_distribution_configuration" "bitcoin-coverage-distribution" {
  provider = aws.compute_region
  name     = "bitcoin-coverage-distribution-${terraform.workspace}"
  distribution {
    region = data.aws_region.compute_region.name
    ami_distribution_configuration {}
  }
}

resource "aws_iam_role" "image_builder_role" {
  provider = aws.compute_region
  name     = "image_builder_role-${terraform.workspace}"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "image_builder_role_policy_ssm" {
  provider   = aws.compute_region
  role       = aws_iam_role.image_builder_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}
resource "aws_iam_role_policy_attachment" "image_builder_role_policy_ec2" {
  provider   = aws.compute_region
  role       = aws_iam_role.image_builder_role.name
  policy_arn = "arn:aws:iam::aws:policy/EC2InstanceProfileForImageBuilder"
}
resource "aws_iam_role_policy_attachment" "image_builder_role_policy_ecr" {
  provider   = aws.compute_region
  role       = aws_iam_role.image_builder_role.name
  policy_arn = "arn:aws:iam::aws:policy/EC2InstanceProfileForImageBuilderECRContainerBuilds"
}


resource "aws_iam_instance_profile" "image_builder_instance_profile" {
  provider = aws.compute_region
  name     = "image_builder_instance_profile-${terraform.workspace}"
  role     = aws_iam_role.image_builder_role.name
}

resource "aws_imagebuilder_infrastructure_configuration" "bitcoin-coverage-configuration" {
  provider              = aws.compute_region
  name                  = "bitcoin-coverage-configuration-${terraform.workspace}"
  instance_profile_name = aws_iam_instance_profile.image_builder_instance_profile.name
  instance_types        = ["c7g.medium"]
  security_group_ids    = [data.aws_security_group.compute_security_group.id]
  subnet_id             = data.aws_subnets.batch_subnets.ids[0]
}

resource "aws_imagebuilder_image" "bitcoin-coverage-ami" {
  provider                         = aws.compute_region
  distribution_configuration_arn   = aws_imagebuilder_distribution_configuration.bitcoin-coverage-distribution.arn
  image_recipe_arn                 = aws_imagebuilder_image_recipe.bitcoin-coverage-recipe.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.bitcoin-coverage-configuration.arn
  timeouts {
    create = "20m"
  }
  tags = {
    Name = "bitcoin-coverage-ami-${terraform.workspace}"
  }
  image_tests_configuration {
    image_tests_enabled = false
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_batch_compute_environment" "bench" {
  provider                        = aws.compute_region
  compute_environment_name_prefix = "bench-${terraform.workspace}-"

  compute_resources {
    max_vcpus          = 32
    security_group_ids = [data.aws_security_group.compute_security_group.id]
    ec2_configuration {
      image_id_override = tolist(aws_imagebuilder_image.bitcoin-coverage-ami.output_resources[0].amis)[0].image
      image_type        = "ECS_AL2"
    }

    subnets = [
      data.aws_subnets.batch_subnets.ids[0],
      data.aws_subnets.batch_subnets.ids[1],
      data.aws_subnets.batch_subnets.ids[2],
    ]

    type                = "SPOT"
    instance_role       = aws_iam_instance_profile.ecs_instance_role.arn
    instance_type       = ["c7g.large"]
    allocation_strategy = "SPOT_PRICE_CAPACITY_OPTIMIZED"
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
        value = "2",
      },
      {
        type  = "MEMORY"
        value = "2048",
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
        value = aws_s3_bucket.bitcoin-coverage-data.id
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
    ]

    executionRoleArn = aws_iam_role.job_role.arn
    jobRoleArn       = aws_iam_role.job_role.arn
  })
  timeout {
    attempt_duration_seconds = 5400
  }
}
