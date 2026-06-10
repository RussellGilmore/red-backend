locals {
  # Tags applied to every resource this module creates.
  # Replaces the embedded provider default_tags, which a shared
  # module must not own — provider configuration belongs to the caller.
  tags = merge(
    {
      Orchestrator = "Terraform"
      Artifact     = "Red-Backend"
      Project      = var.project_name
    },
    var.additional_tags
  )
}
