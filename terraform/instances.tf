# Amazon ECS-Optimized Amazon Linux 
data "aws_ami" "ecs-optimized" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn2-ami-ecs-hvm-*-arm64-ebs"]
  }

  owners = ["amazon"]
}

resource "aws_key_pair" "ssh_key" {
  key_name   = "aureleoules"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEM79mi/xHOtZw+bUfOH8soMjCyO5qOdpLls1tXnR2AD aurele@oules.com"
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

  ami      = data.aws_ami.ecs-optimized.id
  key_name = aws_key_pair.ssh_key.key_name

  ebs_block_device {
    device_name = "/dev/sdf"
    volume_id   = aws_ebs_volume.db.id
  }
}