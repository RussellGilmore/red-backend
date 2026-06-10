# Data sources for current AWS account, partition, and region
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

# Terraform Statefile Bucket
# trivy:ignore:AVD-AWS-0089
resource "aws_s3_bucket" "backend_s3" {
  bucket = "${var.project_name}-s3"

  tags = local.tags
}

# Server-side encryption configuration for the S3 bucket
# SSE-S3 by default; SSE-KMS with bucket key when kms_key_arn is set.
# Justification: CMK encryption is available opt-in via var.kms_key_arn.
# trivy:ignore:AVD-AWS-0132
resource "aws_s3_bucket_server_side_encryption_configuration" "s3_encryption" {
  bucket = aws_s3_bucket.backend_s3.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = var.kms_key_arn != "" ? "aws:kms" : "AES256"
      kms_master_key_id = var.kms_key_arn != "" ? var.kms_key_arn : null
    }

    bucket_key_enabled = var.kms_key_arn != ""
  }
}

# Enable versioning for the S3 bucket
resource "aws_s3_bucket_versioning" "s3_versioning" {
  bucket = aws_s3_bucket.backend_s3.id

  versioning_configuration {
    status = "Enabled"
  }
}

# Enable public access block for the S3 bucket
resource "aws_s3_bucket_public_access_block" "s3_public_access_block" {
  bucket = aws_s3_bucket.backend_s3.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# IAM policy to access the S3 bucket with S3 native locking support
resource "aws_iam_policy" "s3_ddb_policy" {
  name        = "${var.project_name}-Backend-Resource-Policy"
  description = "IAM policy to access the S3 bucket"

  tags = local.tags

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = concat(
      [
        {
          Effect   = "Allow"
          Action   = ["s3:ListBucket"]
          Resource = ["arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.backend_s3.id}"]
        },
        {
          Effect = "Allow"
          Action = [
            "s3:GetObject",
            "s3:PutObject"
          ]
          Resource = ["arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.backend_s3.id}/*"]
        },
        {
          Effect = "Allow"
          Action = [
            "s3:GetObject",
            "s3:PutObject",
            "s3:DeleteObject"
          ]
          Resource = ["arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.backend_s3.id}/*.tflock"]
          Condition = {
            StringEquals = {
              "s3:ExistingObjectTag/terraform-lock" = "true"
            }
          }
        }
      ],
      var.kms_key_arn != "" ? [
        {
          Effect = "Allow"
          Action = [
            "kms:Decrypt",
            "kms:GenerateDataKey"
          ]
          Resource = [var.kms_key_arn]
        }
      ] : []
    )
  })
}
