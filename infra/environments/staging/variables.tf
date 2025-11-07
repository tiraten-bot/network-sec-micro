variable "region" {
  type        = string
  description = "AWS region used for the staging environment"
  default     = "eu-central-1"
}

variable "environment" {
  type        = string
  description = "Logical environment name"
  default     = "staging"
}

variable "account_id" {
  type        = string
  description = "AWS account ID that hosts the staging environment"
}

variable "vpc_cidr" {
  type        = string
  description = "Base CIDR block for the environment VPC"
  default     = "10.60.0.0/16"
}

variable "public_subnets" {
  type        = list(string)
  description = "Public subnet CIDR blocks"
  default     = ["10.60.0.0/20", "10.60.16.0/20", "10.60.32.0/20"]
}

variable "private_subnets" {
  type        = list(string)
  description = "Private subnet CIDR blocks"
  default     = ["10.60.48.0/20", "10.60.64.0/20", "10.60.80.0/20"]
}

variable "grafana_admin_user" {
  type        = string
  description = "Grafana admin username for the monitoring stack"
  default     = "admin"
}

variable "grafana_admin_password" {
  type        = string
  description = "Grafana admin password (use Vault/SSM in real deployments)"
  sensitive   = true
}

