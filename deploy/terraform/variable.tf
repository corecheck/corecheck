variable "aws_access_key_id" {}
variable "aws_secret_access_key" {}
variable "sonar_token" {}

variable "db_user" {}
variable "db_password" {}
variable "db_database" {
  default = "corecheck"
}
variable "ssh_pubkey" {}

variable "dns_name" {
  default = "corecheck.dev"
}

variable "github_token" {}
variable "datadog_api_key" {}

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
