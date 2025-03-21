# Amazon ECS-Optimized Amazon Linux 
data "aws_ami" "ecs-optimized" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn2-ami-ecs-hvm-*-arm64-ebs"]
  }

  owners = ["amazon"]
}

data "aws_ami" "ubuntu_22_04" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-arm64-server-*"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_key_pair" "ssh_key" {
  key_name   = "terraform-${terraform.workspace}"
  public_key = var.ssh_pubkey
}

resource "aws_eip" "lb" {
  instance = aws_instance.db.id
  domain   = "vpc"
}

# create external disk for db data
resource "aws_ebs_volume" "db" {
  availability_zone = "eu-west-3a"
  size              = 10
  type              = "gp2"
  tags = {
    Name = "db"
  }
}

# create security group for db
resource "aws_security_group" "db" {
  name        = "db-${terraform.workspace}"
  description = "Security group for db"

  ingress {
    description = "Postgres"
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "db" {
  instance_type = "t4g.small"

  availability_zone = "eu-west-3a"
  ami               = data.aws_ami.ubuntu_22_04.id
  key_name          = aws_key_pair.ssh_key.key_name
  security_groups = [
    aws_security_group.db.name
  ]

  root_block_device {
    volume_size = 10
  }
}

resource "aws_volume_attachment" "db" {
  device_name = "/dev/sdf"
  volume_id   = aws_ebs_volume.db.id
  instance_id = aws_instance.db.id

  # Provision the DB instance using Ansible. Define instance provisioners here, instead of on
  # the insance, so the EBS volume is guaranteed to be attached.

  # This is a check to ensure SSH comes up before we run the local exec.
  provisioner "remote-exec" {
    inline = ["echo 'Hello World'"]

    connection {
      type = "ssh"
      host = aws_eip.lb.public_ip
      user = "ubuntu"
      private_key = file(var.ssh_private_key_file)
    }
  }

  provisioner "local-exec" {
    command = <<EOF
echo "db ansible_host=${aws_eip.lb.public_ip} ansible_ssh_user=ubuntu" > hosts.ini
EOF
    working_dir = "../ansible"
  }

  provisioner "local-exec" {
    command = "ansible-playbook playbooks/*.yml --private-key ${var.ssh_private_key_file} --ssh-common-args '-o IdentitiesOnly=yes'"
    environment = {
      DB_USER     = var.db_user
      DB_PASSWORD = var.db_password
    }
    working_dir = "../ansible"
  }
}
