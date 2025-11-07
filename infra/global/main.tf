provider "aws" {
  region = var.region

  assume_role {
    role_arn = "arn:aws:iam::${var.account_id}:role/terraform-admin"
  }

  default_tags {
    tags = {
      project     = "network-sec"
      environment = "global"
      managed_by  = "terraform"
    }
  }
}

locals {
  tags = {
    project = "network-sec"
    layer   = "global"
  }
}

# Example placeholder resources. Replace with real infrastructure like
# organizations accounts, shared VPCs, root DNS zones, etc.
# resource "aws_organizations_account" "example" {
#   name  = "shared-services"
#   email = var.shared_services_email
# }
