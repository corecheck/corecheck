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
