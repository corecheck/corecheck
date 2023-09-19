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
  key_name   = "ci"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGuel3J5BthPQnrAjrOqt8lY0X+mU+sx/rUgbB54FVw9 aureleoules@nuflap"
}

resource "aws_eip" "lb" {
  instance = aws_instance.core.id
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

resource "aws_instance" "core" {
  instance_type = "t4g.nano"

  availability_zone = "eu-west-3a"
  ami      = data.aws_ami.ubuntu_22_04.id
  key_name = aws_key_pair.ssh_key.key_name

    root_block_device {
        volume_size = 10
    }
}

resource "aws_volume_attachment" "db" {
  device_name = "/dev/nvme1n1"
  volume_id   = aws_ebs_volume.db.id
  instance_id = aws_instance.core.id
}