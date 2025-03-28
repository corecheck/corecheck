variable "aws_access_key_id" {}
variable "aws_secret_access_key" {
  sensitive = true
}
variable "sonar_token" {
  sensitive = true
}

variable "db_user" {}
variable "db_password" {
  sensitive = true
}
variable "db_database" {
  default = "corecheck"
}
variable "ssh_private_key_file" {}
variable "ssh_pubkey" {}

variable "dns_name" {
  default = "corecheck.dev"
}

variable "github_token" {}
variable "datadog_api_key" {
  sensitive = true
}
