module "api_gateway" {
  source = "./api-gateway"

  s3_bucket   = aws_s3_bucket.corecheck-lambdas-api.id
  db_host     = aws_db_instance.db.address
  db_port     = aws_db_instance.db.port
  db_user     = var.db_user
  db_password = var.db_password
  db_database = var.db_database
  db_sslmode  = var.db_sslmode

  dns_name = var.dns_name

  corecheck_data_bucket_url = "https://${aws_s3_bucket.bitcoin-coverage-data.id}.s3.${aws_s3_bucket.bitcoin-coverage-data.region}.amazonaws.com"

  providers = {
    aws.us_east_1 = aws.us_east_1
  }
}

module "monitoring" {
  source = "./monitoring"

  alert_email                   = var.alert_email
  dns_name                      = var.dns_name
  telegram_bot_token            = var.telegram_bot_token
  telegram_chat_id              = var.telegram_chat_id
  dashboard_compute_region      = data.aws_region.compute_region.name
  public_grafana_admin_user     = var.public_grafana_admin_user
  public_grafana_admin_password = var.public_grafana_admin_password

  providers = {
    aws                = aws
    aws.compute_region = aws.compute_region
  }
}

module "compute" {
  source = "./compute"

  db_host     = aws_db_instance.db.address
  db_port     = aws_db_instance.db.port
  db_user     = var.db_user
  db_password = var.db_password
  db_database = var.db_database
  db_sslmode  = var.db_sslmode

  github_token                 = var.github_token
  corecheck_data_bucket        = aws_s3_bucket.bitcoin-coverage-data.id
  corecheck_data_bucket_url    = "https://${aws_s3_bucket.bitcoin-coverage-data.id}.s3.${aws_s3_bucket.bitcoin-coverage-data.region}.amazonaws.com"
  corecheck_data_bucket_region = aws_s3_bucket.bitcoin-coverage-data.region

  aws_access_key_id     = var.aws_access_key_id
  aws_secret_access_key = var.aws_secret_access_key

  sonar_token                   = var.sonar_token
  telemetry_backend             = var.telemetry_backend
  telemetry_timestream_database = module.monitoring.dashboard_timestream_database_name
  telemetry_timestream_table    = module.monitoring.dashboard_timestream_table_names["dashboard_metrics"]
  telemetry_timestream_region   = data.aws_region.default.name

  lambda_bucket = aws_s3_bucket.corecheck-lambdas.id
  providers = {
    aws.us_east_1      = aws.us_east_1
    aws.compute_region = aws.compute_region
  }
}
