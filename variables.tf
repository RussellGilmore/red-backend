variable "project_name" {
  description = "Project name used for resource naming and the Project tag."
  type        = string
}

variable "additional_tags" {
  description = "Additional tags to apply to all resources created by this module."
  type        = map(string)
  default     = {}
}

variable "kms_key_arn" {
  description = "Optional ARN of a customer-managed KMS key for S3 server-side encryption. When empty (default), SSE-S3 (AES256) is used. When set, SSE-KMS with a bucket key is enabled and the generated IAM policy includes the KMS permissions required for state access."
  type        = string
  default     = ""

  validation {
    condition     = var.kms_key_arn == "" || can(regex("^arn:[^:]+:kms:", var.kms_key_arn))
    error_message = "kms_key_arn must be a valid KMS key ARN if provided."
  }
}
