data "aws_region" "current" {}

data "aws_vpc" "public_grafana" {
  default = true
}

data "aws_subnets" "public_grafana" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.public_grafana.id]
  }
}

data "aws_route53_zone" "public_grafana" {
  name = var.dns_name
}

data "aws_ssm_parameter" "public_grafana_ami" {
  name = "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64"
}

locals {
  public_grafana_subnet_ids = sort(data.aws_subnets.public_grafana.ids)

  public_grafana_bootstrap_bucket_name = "corecheck-grafana-bootstrap-${terraform.workspace}-${data.aws_caller_identity.current.account_id}"

  public_grafana_bootstrap_files = merge(
    {
      "provisioning/datasources/corecheck.yaml" = templatefile("${path.module}/public-grafana-bootstrap/datasources.yaml.tftpl", {
        cloudwatch_name = local.public_grafana_datasource_names.cloudwatch
        default_region  = var.dashboard_compute_region
      })
      "provisioning/dashboards/corecheck.yaml" = templatefile("${path.module}/public-grafana-bootstrap/dashboards.yaml.tftpl", {})
    },
    {
      for key, content in local.provisioned_public_dashboard_templates :
      "dashboards/${key}.json" => content
    }
  )

  public_grafana_bootstrap_revision = sha256(join("", [
    for key in sort(keys(local.public_grafana_bootstrap_files)) :
    "${key}:${local.public_grafana_bootstrap_files[key]}"
  ]))
}

resource "aws_s3_bucket" "public_grafana_bootstrap" {
  bucket        = local.public_grafana_bootstrap_bucket_name
  force_destroy = true
}

resource "aws_s3_bucket_server_side_encryption_configuration" "public_grafana_bootstrap" {
  bucket = aws_s3_bucket.public_grafana_bootstrap.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "public_grafana_bootstrap" {
  bucket = aws_s3_bucket.public_grafana_bootstrap.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_object" "public_grafana_bootstrap" {
  for_each = local.public_grafana_bootstrap_files

  bucket       = aws_s3_bucket.public_grafana_bootstrap.id
  key          = each.key
  content      = each.value
  content_type = endswith(each.key, ".json") ? "application/json" : "text/yaml"
  etag         = md5(each.value)
}

data "aws_iam_policy_document" "public_grafana_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "public_grafana" {
  name               = "${local.dashboard_stack_name}-public-grafana"
  assume_role_policy = data.aws_iam_policy_document.public_grafana_assume_role.json
}

resource "aws_iam_instance_profile" "public_grafana" {
  name = "${local.dashboard_stack_name}-public-grafana"
  role = aws_iam_role.public_grafana.name
}

data "aws_iam_policy_document" "public_grafana_bootstrap" {
  statement {
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.public_grafana_bootstrap.arn]
  }

  statement {
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.public_grafana_bootstrap.arn}/*"]
  }
}

resource "aws_iam_policy" "public_grafana_bootstrap" {
  name   = "${local.dashboard_stack_name}-public-grafana-bootstrap"
  policy = data.aws_iam_policy_document.public_grafana_bootstrap.json
}

resource "aws_iam_role_policy_attachment" "public_grafana_bootstrap" {
  role       = aws_iam_role.public_grafana.name
  policy_arn = aws_iam_policy.public_grafana_bootstrap.arn
}

resource "aws_iam_role_policy_attachment" "public_grafana_cloudwatch" {
  role       = aws_iam_role.public_grafana.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonGrafanaCloudWatchAccess"
}

resource "aws_iam_role_policy_attachment" "public_grafana_ssm" {
  role       = aws_iam_role.public_grafana.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_security_group" "public_grafana_alb" {
  name        = "${local.dashboard_stack_name}-public-grafana-alb"
  description = "Public ingress for the Corecheck Grafana ALB"
  vpc_id      = data.aws_vpc.public_grafana.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "public_grafana_instance" {
  name        = "${local.dashboard_stack_name}-public-grafana-instance"
  description = "Grafana instance traffic from the Corecheck Grafana ALB"
  vpc_id      = data.aws_vpc.public_grafana.id

  ingress {
    from_port       = 3000
    to_port         = 3000
    protocol        = "tcp"
    security_groups = [aws_security_group.public_grafana_alb.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_acm_certificate" "public_grafana" {
  domain_name       = local.public_grafana_domain
  validation_method = "DNS"
}

resource "aws_route53_record" "public_grafana_validation" {
  for_each = {
    for dvo in aws_acm_certificate.public_grafana.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  zone_id = data.aws_route53_zone.public_grafana.zone_id
  name    = each.value.name
  type    = each.value.type
  records = [each.value.record]
  ttl     = 60
}

resource "aws_acm_certificate_validation" "public_grafana" {
  certificate_arn         = aws_acm_certificate.public_grafana.arn
  validation_record_fqdns = [for record in aws_route53_record.public_grafana_validation : record.fqdn]
}

resource "aws_lb" "public_grafana" {
  name               = "cc-grafana-${terraform.workspace}"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.public_grafana_alb.id]
  subnets            = slice(local.public_grafana_subnet_ids, 0, 2)
}

resource "aws_lb_target_group" "public_grafana" {
  name        = "cc-grafana-${terraform.workspace}"
  port        = 3000
  protocol    = "HTTP"
  target_type = "instance"
  vpc_id      = data.aws_vpc.public_grafana.id

  health_check {
    enabled             = true
    path                = "/api/health"
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 30
    matcher             = "200"
  }
}

resource "aws_lb_listener" "public_grafana_http" {
  load_balancer_arn = aws_lb.public_grafana.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb_listener" "public_grafana_https" {
  load_balancer_arn = aws_lb.public_grafana.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = aws_acm_certificate_validation.public_grafana.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.public_grafana.arn
  }
}

resource "aws_instance" "public_grafana" {
  ami                         = data.aws_ssm_parameter.public_grafana_ami.value
  instance_type               = var.public_grafana_instance_type
  subnet_id                   = local.public_grafana_subnet_ids[0]
  vpc_security_group_ids      = [aws_security_group.public_grafana_instance.id]
  iam_instance_profile        = aws_iam_instance_profile.public_grafana.name
  associate_public_ip_address = true
  user_data = templatefile("${path.module}/public-grafana-bootstrap/user_data.sh.tftpl", {
    admin_password     = var.public_grafana_admin_password
    admin_user         = var.public_grafana_admin_user
    aws_region         = data.aws_region.current.name
    bootstrap_bucket   = aws_s3_bucket.public_grafana_bootstrap.id
    bootstrap_revision = local.public_grafana_bootstrap_revision
    grafana_domain     = local.public_grafana_domain
    grafana_image      = var.public_grafana_image
  })
  user_data_replace_on_change = true

  metadata_options {
    http_endpoint = "enabled"
    http_tokens   = "required"
  }

  root_block_device {
    volume_size = 20
  }

  tags = {
    Name = "${local.dashboard_stack_name}-public-grafana"
  }

  depends_on = [aws_s3_object.public_grafana_bootstrap]
}

resource "aws_lb_target_group_attachment" "public_grafana" {
  target_group_arn = aws_lb_target_group.public_grafana.arn
  target_id        = aws_instance.public_grafana.id
  port             = 3000
}

resource "aws_route53_record" "public_grafana" {
  zone_id = data.aws_route53_zone.public_grafana.zone_id
  name    = local.public_grafana_domain
  type    = "A"

  alias {
    name                   = aws_lb.public_grafana.dns_name
    zone_id                = aws_lb.public_grafana.zone_id
    evaluate_target_health = true
  }
}
