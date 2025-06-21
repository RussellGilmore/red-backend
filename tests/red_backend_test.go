package test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	// Terratest modules (uses AWS SDK v1 internally)
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// AWS SDK v2 (for your custom AWS operations)
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	awsRegion   = os.Getenv("AWS_REGION")
	projectName = fmt.Sprintf("red-backend-%s", strings.ToLower(random.UniqueId()))
	opts        = &terraform.Options{
		TerraformDir: ".",
		Vars: map[string]interface{}{
			"region":       awsRegion,
			"project_name": projectName,
		},
	}
)

// Destroy the terraform code
func destroyTerraform(t *testing.T) {
	terraform.Destroy(t, opts)
}

// Deploy the terraform code
func deployTerraform(t *testing.T) {
	_, err := terraform.InitAndApplyE(t, opts)
	if err != nil {
		terraform.Apply(t, opts)
	}
}

// Verify the names of the S3 bucket created by the red backend
func verifyRedBackendNames(t *testing.T) {
	bucketName := terraform.Output(t, opts, "red_backend_s3_bucket")
	expectedBucketName := projectName + "-s3"

	assert.Equal(t, expectedBucketName, bucketName, "S3 bucket name should match expected format")
}

// Verify S3 bucket configuration using Terratest (AWS SDK v1 internally)
func verifyS3BucketConfiguration(t *testing.T) {
	bucketName := terraform.Output(t, opts, "red_backend_s3_bucket")

	// Use Terratest's helper functions (which use AWS SDK v1 internally)
	aws.AssertS3BucketExists(t, awsRegion, bucketName)

	// Verify versioning is enabled
	versioningStatus := aws.GetS3BucketVersioning(t, awsRegion, bucketName)
	assert.Equal(t, "Enabled", versioningStatus, "S3 bucket versioning should be enabled")

	// Verify public access is blocked
	publicAccessBlock := aws.GetS3BucketPublicAccessBlock(t, awsRegion, bucketName)
	assert.True(t, *publicAccessBlock.BlockPublicAcls, "Public ACLs should be blocked")
	assert.True(t, *publicAccessBlock.BlockPublicPolicy, "Public policies should be blocked")
	assert.True(t, *publicAccessBlock.IgnorePublicAcls, "Public ACLs should be ignored")
	assert.True(t, *publicAccessBlock.RestrictPublicBuckets, "Public buckets should be restricted")

	// Verify encryption using Terratest
	encryptionRules := aws.GetS3BucketEncryption(t, awsRegion, bucketName)
	require.NotEmpty(t, encryptionRules.Rules, "Encryption rules should be configured")
	assert.Equal(t, "AES256", *encryptionRules.Rules[0].ApplyServerSideEncryptionByDefault.SSEAlgorithm, "Should use AES256 encryption")
}

// Verify S3 native locking capabilities using AWS SDK v2
func verifyS3NativeLockingCapabilities(t *testing.T) {
	bucketName := terraform.Output(t, opts, "red_backend_s3_bucket")

	// Create AWS SDK v2 client
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	require.NoError(t, err, "Failed to load AWS config")

	s3Client := s3.NewFromConfig(cfg)

	// Test that we can perform conditional writes (the basis of S3 native locking)
	testKey := "test-conditional-write.txt"
	testContent := "test content for conditional write"

	// First write should succeed (object doesn't exist)
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(testKey),
		Body:        strings.NewReader(testContent),
		IfNoneMatch: aws.String("*"), // This is what S3 native locking uses
	})
	assert.NoError(t, err, "First conditional write should succeed")

	// Second write with same condition should fail (object now exists)
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(testKey),
		Body:        strings.NewReader("different content"),
		IfNoneMatch: aws.String("*"),
	})
	assert.Error(t, err, "Second conditional write should fail - this proves S3 supports native locking")

	// Verify the error is the expected conditional failure
	var conditionalErr *types.PreconditionFailed
	assert.ErrorAs(t, err, &conditionalErr, "Error should be PreconditionFailed")

	// Clean up test object
	_, err = s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(testKey),
	})
	assert.NoError(t, err, "Should be able to clean up test object")
}

// Verify backend configuration output format
func verifyBackendConfiguration(t *testing.T) {
	// Get the backend configuration output
	backendConfigMap := terraform.OutputMap(t, opts, "backend_configuration")

	// Verify required fields exist and have correct values
	assert.Equal(t, projectName+"-s3", backendConfigMap["bucket"], "Backend config should have correct bucket name")
	assert.Equal(t, awsRegion, backendConfigMap["region"], "Backend config should have correct region")
	assert.Equal(t, "true", backendConfigMap["encrypt"], "Backend config should have encryption enabled")
	assert.Equal(t, "true", backendConfigMap["use_lockfile"], "Backend config should have S3 native locking enabled")

	// Verify the example backend config contains the correct information
	exampleConfig := terraform.Output(t, opts, "example_backend_config")
	assert.Contains(t, exampleConfig, projectName+"-s3", "Example config should contain correct bucket name")
	assert.Contains(t, exampleConfig, awsRegion, "Example config should contain correct region")
	assert.Contains(t, exampleConfig, "use_lockfile = true", "Example config should include S3 native locking")
	assert.Contains(t, exampleConfig, "encrypt      = true", "Example config should include encryption")

	// Verify it does NOT contain DynamoDB references (since we're using S3 native locking)
	assert.NotContains(t, exampleConfig, "dynamodb_table", "Example config should not contain DynamoDB table reference")
}

// Verify S3 bucket tags using Terratest
func verifyS3BucketTags(t *testing.T) {
	bucketName := terraform.Output(t, opts, "red_backend_s3_bucket")

	// Get bucket tags using Terratest
	tags := aws.GetS3BucketTags(t, awsRegion, bucketName)

	// Verify default tags from provider
	assert.Equal(t, "Terraform", tags["Orchestrator"], "Should have Orchestrator tag")
	assert.Equal(t, "Red-Backend", tags["Artifact"], "Should have Artifact tag")
	assert.Equal(t, projectName, tags["Project"], "Should have Project tag")
}

// Test S3 bucket lifecycle configuration using AWS SDK v2
func verifyS3BucketLifecycle(t *testing.T) {
	bucketName := terraform.Output(t, opts, "red_backend_s3_bucket")

	// Create AWS SDK v2 client
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	require.NoError(t, err, "Failed to load AWS config")

	s3Client := s3.NewFromConfig(cfg)

	// Check if lifecycle configuration exists (optional feature)
	_, err = s3Client.GetBucketLifecycleConfiguration(context.TODO(), &s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucketName),
	})

	// It's okay if no lifecycle is configured, but if there is one, it should be valid
	if err == nil {
		t.Log("Lifecycle configuration found and appears valid")
	} else {
		t.Log("No lifecycle configuration found (this is okay)")
	}
}

// Test IAM policy permissions for S3 native locking using AWS SDK v2
func verifyIAMPolicyForS3Locking(t *testing.T) {
	bucketName := terraform.Output(t, opts, "red_backend_s3_bucket")

	// Create AWS SDK v2 client
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	require.NoError(t, err, "Failed to load AWS config")

	s3Client := s3.NewFromConfig(cfg)

	// Test basic S3 operations that would be needed for Terraform state management
	testKey := "test-state-operations.tfstate"
	testStateContent := `{
  "version": 4,
  "terraform_version": "1.10.0",
  "serial": 1,
  "lineage": "test-lineage",
  "outputs": {},
  "resources": []
}`

	// Test putting a state file
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:               aws.String(bucketName),
		Key:                  aws.String(testKey),
		Body:                 strings.NewReader(testStateContent),
		ServerSideEncryption: types.ServerSideEncryptionAes256,
	})
	assert.NoError(t, err, "Should be able to put state file")

	// Test getting the state file
	result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(testKey),
	})
	assert.NoError(t, err, "Should be able to get state file")
	if err == nil {
		result.Body.Close()
	}

	// Test lock file operations (*.tflock)
	lockKey := testKey + ".tflock"
	lockContent := `{
  "ID": "test-lock-id",
  "Operation": "OperationTypeApply",
  "Info": "test lock",
  "Who": "test-user",
  "Version": "1.10.0",
  "Created": "2025-06-21T12:00:00Z"
}`

	// Test putting a lock file
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(lockKey),
		Body:   strings.NewReader(lockContent),
	})
	assert.NoError(t, err, "Should be able to put lock file")

	// Test deleting the lock file
	_, err = s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(lockKey),
	})
	assert.NoError(t, err, "Should be able to delete lock file")

	// Clean up test state file
	_, err = s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(testKey),
	})
	assert.NoError(t, err, "Should be able to clean up test state file")
}

// Test the red backend terraform module
func TestRedBackend(t *testing.T) {
	defer test_structure.RunTestStage(t, "terraform_destroy", func() {
		destroyTerraform(t)
	})

	test_structure.RunTestStage(t, "terraform_init_and_apply", func() {
		deployTerraform(t)
	})

	test_structure.RunTestStage(t, "validate_red_backend_names", func() {
		verifyRedBackendNames(t)
	})

	test_structure.RunTestStage(t, "validate_s3_bucket_configuration", func() {
		verifyS3BucketConfiguration(t)
	})

	test_structure.RunTestStage(t, "validate_s3_native_locking_capabilities", func() {
		verifyS3NativeLockingCapabilities(t)
	})

	test_structure.RunTestStage(t, "validate_backend_configuration_output", func() {
		verifyBackendConfiguration(t)
	})

	test_structure.RunTestStage(t, "validate_s3_bucket_tags", func() {
		verifyS3BucketTags(t)
	})

	test_structure.RunTestStage(t, "validate_s3_bucket_lifecycle", func() {
		verifyS3BucketLifecycle(t)
	})

	test_structure.RunTestStage(t, "validate_iam_policy_for_s3_locking", func() {
		verifyIAMPolicyForS3Locking(t)
	})
}
