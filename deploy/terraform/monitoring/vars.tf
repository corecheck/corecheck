variable "alert_email" {
  description = "Email address to receive CloudWatch alert notifications"
  type        = string
}

variable "telegram_bot_token" {
  description = "Telegram bot token for alert notifications. Leave empty to disable Telegram alerts."
  type        = string
  default     = ""
  sensitive   = true
}

variable "telegram_chat_id" {
  description = "Telegram chat ID to send alert notifications to. Required if telegram_bot_token is set."
  type        = string
  default     = ""
}

variable "canary_schedule" {
  description = "CloudWatch Synthetics canary schedule expression"
  type        = string
  default     = "rate(30 minutes)"
}

variable "dashboard_compute_region" {
  description = "AWS region that emits Batch and Step Functions metrics for the public dashboards"
  type        = string
}

variable "dns_name" {
  description = "Route53 zone used for public-facing monitoring endpoints"
  type        = string
}

variable "public_grafana_admin_user" {
  description = "Admin username for the self-hosted public Grafana instance"
  type        = string
  default     = "corecheck-admin"
}

variable "public_grafana_admin_password" {
  description = "Admin password for the self-hosted public Grafana instance"
  type        = string
  sensitive   = true
}

variable "public_grafana_instance_type" {
  description = "EC2 instance type for the self-hosted public Grafana instance"
  type        = string
  default     = "t3.small"
}

variable "public_grafana_image" {
  description = "Container image used for the self-hosted public Grafana instance"
  type        = string
  default     = "grafana/grafana-oss:latest"
}
