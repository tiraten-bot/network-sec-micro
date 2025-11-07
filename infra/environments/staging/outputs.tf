output "cluster_name" {
  description = "Name of the EKS cluster deployed in the environment"
  value       = module.kubernetes.cluster_name
}

output "vpc_id" {
  description = "ID of the environment VPC"
  value       = module.network.vpc_id
}

