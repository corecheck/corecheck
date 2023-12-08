# event bridge bus default
data "aws_cloudwatch_event_bus" "default" {
  name = "default"
  provider = aws.compute_region
}
