variable "region" {
  type        = string
  description = "AWS region used for global resources"
  default     = "eu-central-1"
}

variable "account_id" {
  type        = string
  description = "AWS account ID that hosts global resources"
}

variable "shared_services_email" {
  type        = string
  description = "Contact email for shared services account provisioning"
  default     = "shared-services@example.com"
}
