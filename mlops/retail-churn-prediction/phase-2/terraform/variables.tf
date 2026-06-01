variable "aws_region" {
  description = "AWS region for all resources"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Prefix applied to all resource names"
  type        = string
  default     = "retail-churn"
}

variable "ecr_repository_name" {
  description = "ECR repository name for the churn API image"
  type        = string
  default     = "churn-api"
}

variable "ecs_cluster_name" {
  description = "ECS cluster name"
  type        = string
  default     = "churn-cluster"
}

variable "ecs_service_name" {
  description = "ECS service name"
  type        = string
  default     = "churn-api"
}

variable "db_password" {
  description = "Master password for RDS Postgres — supply via TF_VAR_db_password or a .tfvars file"
  type        = string
  sensitive   = true
}

variable "slack_webhook_url" {
  description = "Slack incoming webhook URL for SNS notifications (leave empty to skip)"
  type        = string
  default     = ""
  sensitive   = true
}
