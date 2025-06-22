# Purpose: Define the outputs for the backend S3 bucket
output "red_backend_s3_bucket" {
  value       = aws_s3_bucket.backend_s3.bucket
  description = "The S3 bucket for storing Terraform state files"
}

# Purpose: The backend configuration for easy reference
output "backend_configuration" {
  description = "Backend configuration block for use in other Terraform configurations"
  value = {
    bucket       = aws_s3_bucket.backend_s3.id
    region       = data.aws_region.current.name
    encrypt      = true
    use_lockfile = true
  }
}

# Purpose: Provide an example backend configuration to copy into the terraform block
output "example_backend_config" {
  description = "Example backend configuration to copy into your terraform block"
  value       = <<-EOT
    terraform {
      backend "s3" {
        bucket       = "${aws_s3_bucket.backend_s3.id}"
        key          = "path/to/your/terraform.tfstate"
        region       = "${data.aws_region.current.name}"
        encrypt      = true
        use_lockfile = true
      }
    }
  EOT
}
