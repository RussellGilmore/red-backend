variable "project_name" {
  description = "Set the project name."
  type        = string
  default     = "red-backend-example"
}

variable "region" {
  description = "AWS region to deploy into."
  type        = string
  default     = "us-east-1"
}

# The provider configuration lives with the caller, never inside the module.
provider "aws" {
  region = var.region
}

module "red_backend" {
  source = "../../"

  project_name = var.project_name

  additional_tags = {
    Environment = "example"
  }
}

output "red_backend_s3_bucket" {
  description = "The S3 bucket for storing Terraform state files"
  value       = module.red_backend.red_backend_s3_bucket
}

output "backend_configuration" {
  description = "Backend configuration block for use in other Terraform configurations"
  value       = module.red_backend.backend_configuration
}

output "example_backend_config" {
  description = "Example backend configuration to copy into your terraform block"
  value       = module.red_backend.example_backend_config
}
