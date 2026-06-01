# Single bucket for both DVC artifact storage and MLflow run artifacts.
# Paths:
#   s3://<bucket>/dvc/        — DVC remote (data + model versions)
#   s3://<bucket>/mlflow/     — MLflow artifact store (logs, metrics, plots)

resource "aws_s3_bucket" "artifacts" {
  # Include account ID to guarantee global uniqueness without a random suffix
  bucket        = "${var.project_name}-artifacts-${data.aws_caller_identity.current.account_id}"
  force_destroy = false
}

resource "aws_s3_bucket_versioning" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  rule {
    id     = "expire-old-dvc-noncurrent"
    status = "Enabled"

    filter {
      prefix = "dvc/"
    }

    # Noncurrent (overwritten) DVC objects are kept for 90 days for rollback,
    # then deleted to contain storage costs.
    noncurrent_version_expiration {
      noncurrent_days = 90
    }
  }
}

resource "aws_s3_bucket_public_access_block" "artifacts" {
  bucket                  = aws_s3_bucket.artifacts.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_server_side_encryption_configuration" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
