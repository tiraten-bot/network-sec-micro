package policies.cost

deny[msg] {
  input.resource_changes[_].type == "aws_instance"
  instance = input.resource_changes[_]
  size := instance.change.after.instance_type
  startswith(size, "m5.4xlarge")
  msg = sprintf("Cost policy: %s instance type %s exceeds staging/prod budget", [instance.address, size])
}

