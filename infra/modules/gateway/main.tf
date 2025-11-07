variable "environment" {
  type        = string
  description = "Environment identifier"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID hosting the gateway"
}

variable "public_subnet_ids" {
  type        = list(string)
  description = "Public subnet IDs for load balancers"
}

variable "private_subnet_ids" {
  type        = list(string)
  description = "Private subnet IDs for instances or services"
}

variable "tags" {
  type        = map(string)
  description = "Common tags to apply"
  default     = {}
}

# Placeholder resources for API gateway / ingress stack.
# Extend with load balancers, ECS/Fargate services, or Kubernetes Ingress.

output "example" {
  description = "Placeholder output until gateway resources are defined"
  value       = "${var.environment}-gateway-module"
}

