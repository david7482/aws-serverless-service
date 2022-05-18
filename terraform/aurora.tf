locals {
  aurora_postgresql_name = "${var.app_name}-${var.env}"
}

data "aws_rds_engine_version" "postgresql" {
  engine  = "aurora-postgresql"
  version = "13.6"
}

data "aws_subnets" "public" {
  filter {
    name   = "vpc-id"
    values = [var.vpc_id]
  }

  tags = {
    Purpose = "public-subnet"
  }
}

data "aws_security_group" "default" {
  vpc_id = var.vpc_id
  name   = "default"
}

resource "aws_db_subnet_group" "postgresql" {
  name       = local.aurora_postgresql_name
  subnet_ids = data.aws_subnets.public.ids
}

resource "aws_security_group" "postgresql" {
  name   = "${local.aurora_postgresql_name}-postgresql-sg"
  vpc_id = var.vpc_id

  ingress {
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

  tags = {
    Name = "${local.aurora_postgresql_name}-postgresql-sg"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "random_password" "postgres_password" {
  length = 16
}

resource "aws_ssm_parameter" "postgres_password" {
  name  = "${var.app_name}-postgres-password-${var.env}"
  type  = "SecureString"
  value = random_password.postgres_password.result
}

resource "aws_rds_cluster" "postgresql" {
  cluster_identifier                  = local.aurora_postgresql_name
  database_name                       = "chatbot"
  engine                              = data.aws_rds_engine_version.postgresql.engine
  engine_mode                         = "provisioned"
  engine_version                      = data.aws_rds_engine_version.postgresql.version
  availability_zones                  = ["us-west-2a", "us-west-2b", "us-west-2c"]
  db_subnet_group_name                = aws_db_subnet_group.postgresql.name
  vpc_security_group_ids              = [data.aws_security_group.default.id, aws_security_group.postgresql.id]
  storage_encrypted                   = true
  copy_tags_to_snapshot               = true
  apply_immediately                   = true
  skip_final_snapshot                 = true
  iam_database_authentication_enabled = true
  preferred_backup_window             = "18:00-19:00"
  preferred_maintenance_window        = "wed:20:00-wed:21:00"

  serverlessv2_scaling_configuration {
    max_capacity = 10
    min_capacity = 0.5
  }

  master_username = "postgres"
  master_password = random_password.postgres_password.result

  tags = {
    Name        = local.aurora_postgresql_name
    Environment = var.env
  }
}

resource "aws_rds_cluster_instance" "postgresql_instance" {
  count                        = 1
  identifier                   = "${local.aurora_postgresql_name}-${count.index}"
  cluster_identifier           = aws_rds_cluster.postgresql.id
  instance_class               = "db.serverless"
  engine                       = aws_rds_cluster.postgresql.engine
  engine_version               = aws_rds_cluster.postgresql.engine_version
  db_subnet_group_name         = aws_db_subnet_group.postgresql.name
  publicly_accessible          = true
  apply_immediately            = true
  performance_insights_enabled = true
  tags                         = {
    Name        = local.aurora_postgresql_name
    Environment = var.env
  }
}