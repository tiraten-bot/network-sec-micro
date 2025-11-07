variable "environment" {
  type        = string
  description = "Environment identifier"
}

variable "tags" {
  type        = map(string)
  description = "Common tags to apply"
  default     = {}
}

# Placeholder module for compute resources (e.g., EC2 ASGs, ECS, Lambda).
# Add actual resources as the infrastructure design matures.

output "example" {
  description = "Placeholder output to prevent empty module warnings"
  value       = "${var.environment}-compute-module"
}

