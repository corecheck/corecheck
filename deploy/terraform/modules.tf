# TODO: consider migrating these buckets to direct Terraform variables
#  instead of lookup based on naming convention. Also consider moving
#  toward building the lambda functions directly.

data "aws_s3_bucket" "api_lambdas" {
  bucket = "corecheck-api-lambdas-${terraform.workspace}"
}

data "aws_s3_bucket" "compute_lambdas" {
  bucket = "corecheck-compute-lambdas-${terraform.workspace}"
  provider = aws.compute_region
}

resource "terraform_data" "build_lambdas" {
  triggers_replace = local.function_file_hashes

  provisioner "local-exec" {
    command     = "make build-lambdas"
    working_dir = "../../"
  }
}

module "api_gateway" {
  source = "./api-gateway"

  s3_bucket   = data.aws_s3_bucket.api_lambdas.id
  db_host     = aws_eip.lb.public_ip
  db_port     = 5432
  db_user     = var.db_user
  db_password = var.db_password
  db_database = var.db_database

  dns_name = var.dns_name

  corecheck_data_bucket_url = "https://${aws_s3_bucket.bitcoin-coverage-data.id}.s3.${aws_s3_bucket.bitcoin-coverage-data.region}.amazonaws.com"

  providers = {
    aws.us_east_1 = aws.us_east_1
  }

  depends_on = [ terraform_data.build_lambdas ]
}

module "compute" {
  source = "./compute"

  db_host     = aws_eip.lb.public_ip
  db_port     = 5432
  db_user     = var.db_user
  db_password = var.db_password
  db_database = var.db_database

  github_token          = var.github_token
  corecheck_data_bucket = aws_s3_bucket.bitcoin-coverage-data.id
  corecheck_data_bucket_url = "https://${aws_s3_bucket.bitcoin-coverage-data.id}.s3.${aws_s3_bucket.bitcoin-coverage-data.region}.amazonaws.com"
  corecheck_data_bucket_region = aws_s3_bucket.bitcoin-coverage-data.region

  aws_access_key_id     = var.aws_access_key_id
  aws_secret_access_key = var.aws_secret_access_key

  sonar_token = var.sonar_token
  datadog_api_key = var.datadog_api_key

  lambda_bucket = data.aws_s3_bucket.compute_lambdas.id
  providers = {
    aws.us_east_1 = aws.us_east_1
    aws.compute_region = aws.compute_region
  }

  # Wait for database to be provisioned.
  depends_on = [
    aws_volume_attachment.db,
    terraform_data.build_lambdas
  ]
}
