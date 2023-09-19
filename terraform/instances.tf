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
    key_name = "aureleoules"
    public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEM79mi/xHOtZw+bUfOH8soMjCyO5qOdpLls1tXnR2AD aurele@oules.com"
}

resource "aws_instance" "core" {
  instance_type               = "t4g.nano"
  associate_public_ip_address = true

  ami = data.aws_ami.ecs-optimized.id
  key_name = aws_key_pair.ssh_key.key_name
}