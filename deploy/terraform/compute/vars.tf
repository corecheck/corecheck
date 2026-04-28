variable "db_host" {
  type        = string
  description = "db host"
}

variable "db_port" {
  type        = number
  description = "db port"
}

variable "db_user" {
  type        = string
  description = "db user"
}

variable "db_password" {
  type        = string
  description = "db password"
}

variable "db_database" {
  type        = string
  description = "db database"
}

variable "db_sslmode" {
  type        = string
  description = "db ssl mode"
}

variable "github_token" {
  type        = string
  description = "github token"
}

variable "corecheck_data_bucket" {
  type        = string
  description = "corecheck data bucket"
}
variable "corecheck_data_bucket_region" {
  type        = string
  description = "corecheck data bucket"
}

variable "corecheck_data_bucket_url" {
  type        = string
  description = "corecheck data bucket url"
}

variable "aws_access_key_id" {
  type        = string
  description = "aws access key id"
}

variable "aws_secret_access_key" {
  type        = string
  description = "aws secret access key"
}

variable "sonar_token" {
  type        = string
  description = "sonar token"
}

variable "lambda_bucket" {
  type        = string
  description = "lambda bucket"
}

variable "telemetry_backend" {
  type        = string
  description = "telemetry backend used by compute workloads"
  default     = "timestream"

  validation {
    condition     = contains(["datadog", "timestream"], var.telemetry_backend)
    error_message = "telemetry_backend must be either datadog or timestream."
  }
}

variable "telemetry_timestream_database" {
  type        = string
  description = "timestream database name for telemetry metrics"
  default     = ""
}

variable "telemetry_timestream_table" {
  type        = string
  description = "timestream table name for telemetry metrics"
  default     = ""
}

variable "telemetry_timestream_region" {
  type        = string
  description = "AWS region for the timestream telemetry backend"
  default     = ""
}
