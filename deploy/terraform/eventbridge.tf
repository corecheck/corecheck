# event bridge bus
resource "aws_cloudwatch_event_bus" "eventbridge" {
  name = "corecheck-events"
  provider = aws.compute_region
}
