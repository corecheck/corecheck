resource "aws_security_group" "rds" {
  name        = "db-rds-${terraform.workspace}"
  description = "Security group for RDS Postgres"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    description = "Postgres"
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_subnet_group" "db" {
  name       = "corecheck-${terraform.workspace}"
  subnet_ids = data.aws_subnets.example.ids

  tags = {
    Name = "corecheck-${terraform.workspace}"
  }
}

resource "aws_db_instance" "db" {
  identifier             = "corecheck-${terraform.workspace}"
  engine                 = "postgres"
  instance_class         = "db.t4g.small"
  allocated_storage      = 20
  max_allocated_storage  = 100
  storage_type           = "gp3"
  storage_encrypted      = true
  publicly_accessible    = true
  db_subnet_group_name   = aws_db_subnet_group.db.name
  vpc_security_group_ids = [aws_security_group.rds.id]

  db_name  = var.db_database
  username = var.db_user
  password = var.db_password
  port     = 5432

  backup_retention_period = 7
  copy_tags_to_snapshot   = true
  deletion_protection     = true
  skip_final_snapshot     = true
  apply_immediately       = true

  tags = {
    Name = "corecheck-${terraform.workspace}"
  }
}
