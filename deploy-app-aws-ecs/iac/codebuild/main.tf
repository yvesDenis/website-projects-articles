terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.0"
    }
  }
}

resource "aws_s3_bucket" "base_app_codebuild_bucket" {
  bucket = "base-app-codebuild-bucket"
}

data "aws_iam_role" "base_app_codebuild_role" {
  name = "codebuild-codeBuildBaseApp-service-role"
}

resource "aws_codebuild_project" "base_app_codebuild" {
  name          = "base-app-codebuild"
  description   = "Codebuild project for testing base-app"
  build_timeout = "5"
  service_role  = data.aws_iam_role.base_app_codebuild_role.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  cache {
    type     = "S3"
    location = aws_s3_bucket.base_app_codebuild_bucket.bucket
  }

  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "aws/codebuild/standard:3.0"
    type                        = "LINUX_CONTAINER"
    image_pull_credentials_type = "CODEBUILD"
  }

  logs_config {
    cloudwatch_logs {
      group_name  = "log-group"
      stream_name = "log-stream"
    }

    s3_logs {
      status   = "ENABLED"
      location = "${aws_s3_bucket.base_app_codebuild_bucket.id}/test-log"
    }
  }

  source {
    type            = "GITHUB"
    location        = "https://github.com/yvesDenis/website-projects-articles.git"
    git_clone_depth = 5

    git_submodules_config {
      fetch_submodules = true
    }
  }

  source_version = "master"

  tags = {
    name = "base-app"
    subject = "github"
  }
}