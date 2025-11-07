package policies.security

deny[msg] {
  sg := input.resource_changes[_]
  sg.type == "aws_security_group"
  rule := sg.change.after.ingress[_]
  rule.cidr_blocks[_] == "0.0.0.0/0"
  msg = sprintf("Security policy: %s allows public ingress; tighten CIDR", [sg.address])
}

