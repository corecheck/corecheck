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
data "aws_security_group" "compute_security_group" {
  provider = aws.compute_region
  name     = "default"
}

resource "aws_imagebuilder_component" "bitcoin-coverage-component" {
  provider = aws.compute_region
  name     = "CPUAffinity"
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
  name         = "bitcoin-coverage-recipe"
  version      = "1.0.7"
  parent_image = data.aws_ami.amazon-linux-2.id
  block_device_mapping {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = false
      volume_size           = 30
      volume_type           = "io2"
      iops = 10000
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
  name     = "bitcoin-coverage-distribution"
  distribution {
    region = data.aws_region.compute_region.name
    ami_distribution_configuration {}
  }
}

resource "aws_iam_role" "image_builder_role" {
  provider = aws.compute_region
  name     = "image_builder_role"
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
  name     = "image_builder_instance_profile"
  role     = aws_iam_role.image_builder_role.name
}

resource "aws_imagebuilder_infrastructure_configuration" "bitcoin-coverage-configuration" {
  provider              = aws.compute_region
  name                  = "bitcoin-coverage-configuration"
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
    Name = "bitcoin-coverage-ami"
  }
  image_tests_configuration {
    image_tests_enabled = false
  }

  lifecycle {
    create_before_destroy = true
  }
}


resource "aws_batch_compute_environment" "jobs_compute" {
  provider                        = aws.compute_region
  compute_environment_name_prefix = "jobs_compute"

  compute_resources {
    max_vcpus          = 32
    security_group_ids = [data.aws_security_group.compute_security_group.id]
    ec2_configuration {
      image_id_override = tolist(aws_imagebuilder_image.bitcoin-coverage-ami.output_resources[0].amis)[0].image
      image_type        = "ECS_AL2"
    }

    subnets = [
      data.aws_subnets.batch_subnets.ids[0],
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
      }
    ]

    executionRoleArn = aws_iam_role.job_role.arn
    jobRoleArn       = aws_iam_role.job_role.arn
  })
  timeout {
    attempt_duration_seconds = 5400
  }
}
