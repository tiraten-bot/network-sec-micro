variable "environment" {
  type        = string
  description = "Environment identifier"
}

variable "vpc_id" {
  type        = string
  description = "VPC to attach security resources to"
}

variable "tags" {
  type        = map(string)
  description = "Common tags to apply"
  default     = {}
}

# Default security group denies all inbound/outbound; customize per service.
resource "aws_security_group" "default_deny" {
  name        = "${var.environment}-default-deny"
  description = "Default deny-all security group"
  vpc_id      = var.vpc_id

  revoke_rules_on_delete = true

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = []
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = []
  }

  tags = merge(var.tags, {
    Name = "${var.environment}-default-deny"
  })
}

output "default_security_group_id" {
  description = "ID of the default deny security group"
  value       = aws_security_group.default_deny.id
}

