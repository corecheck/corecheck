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

variable "dns_name" {
    type        = string
    description = "dns name"
}

variable "s3_bucket" {
    type        = string
    description = "s3 bucket"
}

variable "corecheck_data_bucket_url" {
  type        = string
  description = "corecheck data bucket url"
}
