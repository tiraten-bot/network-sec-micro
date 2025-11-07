terraform {
  backend "s3" {
    bucket         = "network-sec-terraform-state"
    key            = "global/terraform.tfstate"
    region         = "eu-central-1"
    dynamodb_table = "network-sec-terraform-locks"
    encrypt        = true
  }
}
