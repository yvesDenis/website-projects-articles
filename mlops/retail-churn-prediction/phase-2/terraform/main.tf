terraform {
  required_version = ">= 1.9"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # After `terraform apply` creates the S3 bucket, migrate state to it:
  #   1. Uncomment this backend block and fill in the bucket name.
  #   2. Run `terraform init -migrate-state`.
  #
  # backend "s3" {
  #   bucket = "<your-state-bucket>"
  #   key    = "retail-churn/phase-2/terraform.tfstate"
  #   region = "us-east-1"
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project   = "retail-churn"
      Phase     = "2-aws"
      ManagedBy = "terraform"
    }
  }
}
