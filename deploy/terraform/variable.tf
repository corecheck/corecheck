variable "aws_access_key_id" {}
variable "aws_secret_access_key" {}
variable "sonar_token" {}

variable "db_user" {}
variable "db_password" {}
variable "db_database" {
  default = "corecheck"
}

variable "github_token" {}
variable "datadog_api_key" {}
