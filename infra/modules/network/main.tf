variable "environment" {
  type        = string
  description = "Environment identifier (dev/staging/prod)"
}

variable "vpc_cidr" {
  type        = string
  description = "CIDR block for the VPC"
}

variable "public_subnets" {
  type        = list(string)
  description = "CIDR blocks for public subnets"
}

variable "private_subnets" {
  type        = list(string)
  description = "CIDR blocks for private subnets"
}

variable "tags" {
  type        = map(string)
  description = "Common tags to apply to resources"
  default     = {}
}

resource "aws_vpc" "this" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = merge(var.tags, {
    Name = "${var.environment}-network-sec-vpc"
  })
}

resource "aws_subnet" "public" {
  for_each = { for idx, cidr in var.public_subnets : idx => cidr }

  vpc_id            = aws_vpc.this.id
  cidr_block        = each.value
  map_public_ip_on_launch = true

  tags = merge(var.tags, {
    Name = "${var.environment}-public-${each.key}"
    Tier = "public"
  })
}

resource "aws_subnet" "private" {
  for_each = { for idx, cidr in var.private_subnets : idx => cidr }

  vpc_id     = aws_vpc.this.id
  cidr_block = each.value

  tags = merge(var.tags, {
    Name = "${var.environment}-private-${each.key}"
    Tier = "private"
  })
}

output "vpc_id" {
  description = "ID of the created VPC"
  value       = aws_vpc.this.id
}

output "public_subnet_ids" {
  description = "IDs of the public subnets"
  value       = [for subnet in aws_subnet.public : subnet.id]
}

output "private_subnet_ids" {
  description = "IDs of the private subnets"
  value       = [for subnet in aws_subnet.private : subnet.id]
}

