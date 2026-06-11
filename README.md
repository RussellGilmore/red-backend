# Red Backend

## [![Red Backend Module](https://github.com/RussellGilmore/red-backend/actions/workflows/module-test.yml/badge.svg?branch=main)](https://github.com/RussellGilmore/red-backend/actions/workflows/module-test.yml)

**Requirements:**

1. Terraform >= 1.15.0
2. Trivy >= 0.68.2

Trivy can be installed via Homebrew on macOS with the command:

```bash
brew install aquasecurity/trivy/trivy
```

## Security posture

-   All public access blocked; versioning always on
-   SSE-S3 encryption by default, optional customer-managed KMS via
    `kms_key_arn` (IAM policy picks up the required KMS permissions
    automatically)
-   Generated IAM policy scoped to the single state bucket, with lock-file
    permissions gated by object tag
-   Code scanned with Trivy and gitleaks on every commit; integration-tested
    with Terratest against a live AWS account

## Usage

See [`examples/complete`](./examples/complete) for a working configuration. The
module does not configure the AWS provider — that belongs to you:

```hcl
provider "aws" {
  region = "us-east-1"
}

module "backend" {
  source = "RussellGilmore/red-backend/aws"

  project_name = "my-project"
}
```

<!-- prettier-ignore-start -->
<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.15.0 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | >= 6.47.0 |

## Providers

| Name | Version |
| ---- | ------- |
| <a name="provider_aws"></a> [aws](#provider\_aws) | >= 6.47.0 |

## Modules

No modules.

## Resources

| Name | Type |
| ---- | ---- |
| [aws_iam_policy.s3_ddb_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_s3_bucket.backend_s3](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket) | resource |
| [aws_s3_bucket_public_access_block.s3_public_access_block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_public_access_block) | resource |
| [aws_s3_bucket_server_side_encryption_configuration.s3_encryption](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_server_side_encryption_configuration) | resource |
| [aws_s3_bucket_versioning.s3_versioning](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_versioning) | resource |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
| [aws_partition.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/partition) | data source |
| [aws_region.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region) | data source |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_additional_tags"></a> [additional\_tags](#input\_additional\_tags) | Additional tags to apply to all resources created by this module. | `map(string)` | `{}` | no |
| <a name="input_kms_key_arn"></a> [kms\_key\_arn](#input\_kms\_key\_arn) | Optional ARN of a customer-managed KMS key for S3 server-side encryption. When empty (default), SSE-S3 (AES256) is used. When set, SSE-KMS with a bucket key is enabled and the generated IAM policy includes the KMS permissions required for state access. | `string` | `""` | no |
| <a name="input_project_name"></a> [project\_name](#input\_project\_name) | Project name used for resource naming and the Project tag. | `string` | n/a | yes |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_backend_configuration"></a> [backend\_configuration](#output\_backend\_configuration) | Backend configuration block for use in other Terraform configurations |
| <a name="output_example_backend_config"></a> [example\_backend\_config](#output\_example\_backend\_config) | Example backend configuration to copy into your terraform block |
| <a name="output_red_backend_s3_bucket"></a> [red\_backend\_s3\_bucket](#output\_red\_backend\_s3\_bucket) | The S3 bucket for storing Terraform state files |
<!-- END_TF_DOCS -->
<!-- prettier-ignore-end -->
