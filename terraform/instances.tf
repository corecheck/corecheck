resource "aws_instance" "core" {
  instance_type               = "t4g.nano"
  associate_public_ip_address = true
}