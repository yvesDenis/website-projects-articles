locals {
  tfc_hostname     = "app.terraform.io" 
  tfc_organization = "iac-projects"
  workspace        = "base-app-iac"
  region           = "ca-central-1"
}

generate "remote_state" {
  path      = "backend.tf"
  if_exists = "overwrite_terragrunt"
  contents = <<EOF
  terraform {
    required_providers {
        aws = {
        source  = "hashicorp/aws"
        version = ">= 4.5"
        }
    }
    backend "remote" {
        hostname = "${local.tfc_hostname}"
        organization = "${local.tfc_organization}"
        workspaces {
        name = "${local.workspace}"
        }
    }
}
EOF
}

generate "provider" {
  path = "provider.tf"
  if_exists = "overwrite_terragrunt"
  contents = <<EOF
provider "aws" {
  region = "${local.region}"
}
EOF
}