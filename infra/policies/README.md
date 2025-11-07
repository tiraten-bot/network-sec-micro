# Policy as Code

This directory contains security and cost guardrails enforced during the
Terraform plan/apply workflow. Policies can be evaluated with Open Policy Agent
(OPA), Conftest, or Terraform Cloud/Enterprise Sentinel.

- `cost/`: Policies that prevent unexpectedly expensive resources.
- `security/`: Policies enforcing encryption, restricted CIDRs, required tags, etc.

Policies should fail CI pipelines on violations of high/critical severity rules.

