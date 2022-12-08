terraform {
    source = "${get_terragrunt_dir()}/../modules//codebuild"
}

include "root" {
  path = find_in_parent_folders()
}