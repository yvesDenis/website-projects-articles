resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-db-subnet-group"
  subnet_ids = aws_subnet.private[*].id
}

resource "aws_db_instance" "postgres" {
  identifier     = "${var.project_name}-postgres"
  engine         = "postgres"
  engine_version = "16"

  # db.t4g.micro: ARM-based, ~$12/month — cheapest option for Postgres on RDS.
  # Upgrade to db.t3.small if you need more than ~20 concurrent connections.
  instance_class    = "db.t4g.micro"
  allocated_storage = 20
  storage_type      = "gp2"

  db_name  = "churn"
  username = "churn"
  password = var.db_password

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]

  multi_az            = false
  publicly_accessible = false

  # Safe to skip for dev; enable for production to prevent accidental deletion
  skip_final_snapshot = true
  deletion_protection = false

  backup_retention_period = 7
  backup_window           = "03:00-04:00"
  maintenance_window      = "Mon:04:00-Mon:05:00"
}

# Store the full DATABASE_URL in SSM so ECS can inject it as a secret without
# it appearing in the task definition JSON (which is visible in the AWS console).
resource "aws_ssm_parameter" "db_url" {
  name  = "/churn/db-url"
  type  = "SecureString"
  value = "postgresql://churn:${var.db_password}@${aws_db_instance.postgres.address}:${aws_db_instance.postgres.port}/churn"
}
