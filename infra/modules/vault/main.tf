variable "environment" {
  type        = string
  description = "Environment identifier"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID where Vault will be deployed"
}

variable "tags" {
  type        = map(string)
  description = "Common tags to apply"
  default     = {}
}

# Placeholder for Vault deployment (EC2/AutoScaling or ECS/Kubernetes).
# Use this module to provision security groups, load balancers, and IAM roles
# required for Vault servers.

output "example" {
  description = "Placeholder output until Vault resources are added"
  value       = "${var.environment}-vault-module"
}

