output "cluster_name" {
  description = "Name of the EKS cluster deployed in the environment"
  value       = module.kubernetes.cluster_name
}

output "vpc_id" {
  description = "ID of the environment VPC"
  value       = module.network.vpc_id
}

output "postgres_endpoint" {
  description = "Endpoint for the PostgreSQL database"
  value       = module.compute.postgres_endpoint
}

output "redis_primary_endpoint" {
  description = "Primary Redis endpoint"
  value       = module.compute.redis_primary_endpoint
}

output "mongo_endpoint" {
  description = "DocumentDB endpoint"
  value       = module.compute.mongo_endpoint
}

output "kafka_bootstrap_brokers" {
  description = "Kafka bootstrap brokers"
  value       = module.compute.kafka_bootstrap_brokers
}

