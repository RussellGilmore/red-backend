variable "project_name" {
  description = "Set the project name."
  type        = string
}

variable "region" {
  description = "Set the appropriate AWS region."
  type        = string
}

module "red-backend" {
  source = "../red-backend"

  project_name = var.project_name
  region       = var.region
}

output "red_backend_s3_bucket" {
  value       = module.red-backend.red_backend_s3_bucket
  description = "The S3 bucket for storing Terraform state files"
}

output "red_backend_ddb_table" {
  value       = module.red-backend.red_backend_ddb_table
  description = "The DynamoDB table for storing Terraform state lock status"
}
