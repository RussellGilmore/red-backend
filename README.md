# Red Backend

**Requirements:**

1. Terraform 1.12.1
2. Trivy 0.63.0

[![Red Backend Module](

## [![Red Backend Module](https://github.com/RussellGilmore/red-backend/actions/workflows/module-test.yml/badge.svg?branch=main)](https://github.com/RussellGilmore/red-backend/actions/workflows/module-test.yml)

A simple remote backend module for AWS Terraform creations.

<!-- prettier-ignore-start -->
<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | 1.12.1 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | 6.0.0-beta1 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | 6.0.0-beta1 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_dynamodb_table.ddb_lock_status_table](https://registry.terraform.io/providers/hashicorp/aws/6.0.0-beta1/docs/resources/dynamodb_table) | resource |
| [aws_iam_policy.s3_ddb_policy](https://registry.terraform.io/providers/hashicorp/aws/6.0.0-beta1/docs/resources/iam_policy) | resource |
| [aws_s3_bucket.backend_s3](https://registry.terraform.io/providers/hashicorp/aws/6.0.0-beta1/docs/resources/s3_bucket) | resource |
| [aws_s3_bucket_public_access_block.s3_public_access_block](https://registry.terraform.io/providers/hashicorp/aws/6.0.0-beta1/docs/resources/s3_bucket_public_access_block) | resource |
| [aws_s3_bucket_server_side_encryption_configuration.s3_encryption](https://registry.terraform.io/providers/hashicorp/aws/6.0.0-beta1/docs/resources/s3_bucket_server_side_encryption_configuration) | resource |
| [aws_s3_bucket_versioning.s3_versioning](https://registry.terraform.io/providers/hashicorp/aws/6.0.0-beta1/docs/resources/s3_bucket_versioning) | resource |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/6.0.0-beta1/docs/data-sources/caller_identity) | data source |
| [aws_partition.current](https://registry.terraform.io/providers/hashicorp/aws/6.0.0-beta1/docs/data-sources/partition) | data source |
| [aws_region.current](https://registry.terraform.io/providers/hashicorp/aws/6.0.0-beta1/docs/data-sources/region) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_project_name"></a> [project\_name](#input\_project\_name) | Set the project name. | `string` | `"red-test"` | no |
| <a name="input_region"></a> [region](#input\_region) | Set the appropriate AWS region. | `string` | `"us-east-1"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_red_backend_ddb_table"></a> [red\_backend\_ddb\_table](#output\_red\_backend\_ddb\_table) | The DynamoDB table for storing Terraform state lock status |
| <a name="output_red_backend_s3_bucket"></a> [red\_backend\_s3\_bucket](#output\_red\_backend\_s3\_bucket) | The S3 bucket for storing Terraform state files |
<!-- END_TF_DOCS -->
<!-- prettier-ignore-end -->
