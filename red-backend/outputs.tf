# Purpose: Define the outputs for the backend S3 bucket
output "red_backend_s3_bucket" {
  value       = aws_s3_bucket.backend_s3.bucket
  description = "The S3 bucket for storing Terraform state files"
}

# Purpose: Define the outputs for the DynamoDB table
output "red_backend_ddb_table" {
  value       = aws_dynamodb_table.ddb_lock_status_table.name
  description = "The DynamoDB table for storing Terraform state lock status"
}
