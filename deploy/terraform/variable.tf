variable "aws_access_key_id" {}
variable "aws_secret_access_key" {}
variable "sonar_token" {}

variable "db_user" {}
variable "db_password" {}
variable "db_database" {
  default = "corecheck"
}
variable "db_sslmode" {
  default = "require"
}

variable "dns_name" {
  default = "corecheck.dev"
}

variable "github_token" {}

variable "telemetry_backend" {
  description = "Telemetry backend for compute workloads"
  type        = string
  default     = "timestream"

  validation {
    condition     = contains(["datadog", "timestream"], var.telemetry_backend)
    error_message = "telemetry_backend must be either datadog or timestream."
  }
}

variable "alert_email" {
  description = "Email address to receive CloudWatch alert notifications"
  type        = string
}

variable "telegram_bot_token" {
  description = "Telegram bot token for alert notifications. Leave empty to disable."
  type        = string
  default     = ""
  sensitive   = true
}

variable "telegram_chat_id" {
  description = "Telegram chat ID to send alert notifications to."
  type        = string
  default     = ""
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
