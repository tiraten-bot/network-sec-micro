# Infrastructure as Code (Terraform)

This directory contains the Terraform-based infrastructure definition for the Network Security Microservices platform. It is organized into reusable modules, environment compositions, global foundations, policies, automation scripts, and CI/CD pipelines.

## Layout

- `global/`: Org-wide resources that are applied once (state backends, shared networking, IAM bootstrap).
- `environments/{dev,staging,prod}/`: Environment-specific stacks that compose modules.
- `modules/`: Reusable building blocks (network, security, monitoring, etc.).
- `policies/`: Policy-as-code rules for security and cost guardrails.
- `scripts/`: Helper scripts for local automation (fmt, validate, plan, apply, drift detection).
- `ci/`: CI/CD pipeline definitions for Terraform plan/apply workflows.

See inline documentation within each module or environment for details on inputs, outputs, and usage.
