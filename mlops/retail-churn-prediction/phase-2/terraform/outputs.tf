output "alb_url" {
  description = "Public URL of the Application Load Balancer"
  value       = "http://${aws_lb.main.dns_name}"
}

output "ecr_repository_url" {
  description = "ECR repository URL — use this to tag and push Docker images"
  value       = aws_ecr_repository.churn_api.repository_url
}

output "s3_bucket_name" {
  description = "S3 bucket name for DVC remote and MLflow artifacts"
  value       = aws_s3_bucket.artifacts.bucket
}

output "rds_endpoint" {
  description = "RDS Postgres host:port — used in DATABASE_URL"
  value       = aws_db_instance.postgres.endpoint
  sensitive   = true
}

output "drift_sns_topic_arn" {
  description = "SNS topic ARN that drift_report.py publishes to when drift is detected"
  value       = aws_sns_topic.drift_alerts.arn
}

output "alerts_sns_topic_arn" {
  description = "SNS topic ARN that CloudWatch alarms publish to"
  value       = aws_sns_topic.alerts.arn
}

output "dvc_remote_url" {
  description = "DVC remote URL — paste this into .dvc/config or run: make dvc-configure"
  value       = "s3://${aws_s3_bucket.artifacts.bucket}/dvc"
}
