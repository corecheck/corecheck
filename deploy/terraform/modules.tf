module "api_gateway" {
  source = "./api_gateway"

  s3_bucket   = aws_s3_bucket.corecheck-lambdas-api.id
  db_host     = aws_instance.db.public_ip
  db_port     = 5432
  db_user     = var.db_user
  db_password = var.db_password
  db_database = var.db_database
}
