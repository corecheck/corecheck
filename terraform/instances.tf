# Amazon ECS-Optimized Amazon Linux 
data "aws_ami" "ecs-optimized" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn2-ami-ecs-hvm-*-x86_64-ebs"]
  }

  owners = ["amazon"]
}

resource "aws_instance" "core" {
  instance_type               = "t4g.nano"
  associate_public_ip_address = true

  ami = data.aws_ami.ecs-optimized.id
}