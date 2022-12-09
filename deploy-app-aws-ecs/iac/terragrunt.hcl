locals {
  tfc_hostname     = "app.terraform.io" 
  tfc_organization = "iac-projects"
  workspace        = "base-app"
  region           = "ca-central-1"
}

generate "remote_state" {
  path      = "backend.tf"
  if_exists = "overwrite_terragrunt"
  contents = <<EOF
  terraform {
      backend "remote" {
      hostname = "${local.tfc_hostname}"
      organization = "${local.tfc_organization}"
      workspaces {
        name = "${local.workspace}-${path_relative_to_include()}"
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
  access_key = "${get_env("AWS_ACCESS_KEY_ID", "access_key")}"
  secret_key = "${get_env("AWS_SECRET_ACCESS_KEY", "secret_key")}"
  token = "${get_env("AWS_SESSION_TOKEN", "token")}"
}
EOF
}