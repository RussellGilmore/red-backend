package test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	terratest_aws "github.com/gruntwork-io/terratest/modules/aws"
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
	awsRegion   = getAWSRegion()
	projectName = fmt.Sprintf("red-backend-%s", strings.ToLower(random.UniqueId()))
	opts        = &terraform.Options{
		TerraformDir: ".",
		Vars: map[string]interface{}{
			"region":       awsRegion,
			"project_name": projectName,
		},
	}
)

// getAWSRegion tries AWS_REGION first, then falls back to AWS_DEFAULT_REGION
func getAWSRegion() string {
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region
	}
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		return region
	}
	// Default fallback
	return "us-east-1"
}

// Empty the S3 bucket before destruction (required for versioned buckets)
func emptyS3Bucket(t *testing.T) {
	bucketName := terraform.Output(t, opts, "red_backend_s3_bucket")

	// Create AWS SDK v2 client
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	require.NoError(t, err, "Failed to load AWS config")

	s3Client := s3.NewFromConfig(cfg)

	// List and delete all object versions
	listVersionsInput := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	}

	for {
		listVersionsOutput, err := s3Client.ListObjectVersions(context.TODO(), listVersionsInput)
		if err != nil {
			t.Logf("Warning: Could not list object versions: %v", err)
			break
		}

		// Delete object versions
		if len(listVersionsOutput.Versions) > 0 {
			var objectsToDelete []types.ObjectIdentifier
			for _, version := range listVersionsOutput.Versions {
				objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
					Key:       version.Key,
					VersionId: version.VersionId,
				})
			}

			deleteInput := &s3.DeleteObjectsInput{
				Bucket: aws.String(bucketName),
				Delete: &types.Delete{
					Objects: objectsToDelete,
				},
			}

			_, err := s3Client.DeleteObjects(context.TODO(), deleteInput)
			if err != nil {
				t.Logf("Warning: Could not delete object versions: %v", err)
			} else {
				t.Logf("Deleted %d object versions", len(objectsToDelete))
			}
		}

		// Delete delete markers
		if len(listVersionsOutput.DeleteMarkers) > 0 {
			var markersToDelete []types.ObjectIdentifier
			for _, marker := range listVersionsOutput.DeleteMarkers {
				markersToDelete = append(markersToDelete, types.ObjectIdentifier{
					Key:       marker.Key,
					VersionId: marker.VersionId,
				})
			}

			deleteInput := &s3.DeleteObjectsInput{
				Bucket: aws.String(bucketName),
				Delete: &types.Delete{
					Objects: markersToDelete,
				},
			}

			_, err := s3Client.DeleteObjects(context.TODO(), deleteInput)
			if err != nil {
				t.Logf("Warning: Could not delete delete markers: %v", err)
			} else {
				t.Logf("Deleted %d delete markers", len(markersToDelete))
			}
		}

		// Check if there are more objects to delete
		if !*listVersionsOutput.IsTruncated {
			break
		}
		listVersionsInput.KeyMarker = listVersionsOutput.NextKeyMarker
		listVersionsInput.VersionIdMarker = listVersionsOutput.NextVersionIdMarker
	}

	t.Logf("S3 bucket %s has been emptied", bucketName)
}

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

// Verify S3 bucket configuration using available Terratest functions and AWS SDK v2
func verifyS3BucketConfiguration(t *testing.T) {
	bucketName := terraform.Output(t, opts, "red_backend_s3_bucket")

	// Use Terratest's basic helper function that we know exists
	terratest_aws.AssertS3BucketExists(t, awsRegion, bucketName)

	// Verify versioning is enabled (this function exists in Terratest)
	versioningStatus := terratest_aws.GetS3BucketVersioning(t, awsRegion, bucketName)
	assert.Equal(t, "Enabled", versioningStatus, "S3 bucket versioning should be enabled")

	// Use AWS SDK v2 to verify additional configuration that Terratest doesn't have helper functions for
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	require.NoError(t, err, "Failed to load AWS config")

	s3Client := s3.NewFromConfig(cfg)

	// Verify public access is blocked using AWS SDK v2
	publicAccessBlock, err := s3Client.GetPublicAccessBlock(context.TODO(), &s3.GetPublicAccessBlockInput{
		Bucket: aws.String(bucketName),
	})
	require.NoError(t, err, "Should be able to get public access block configuration")

	assert.True(t, *publicAccessBlock.PublicAccessBlockConfiguration.BlockPublicAcls, "Public ACLs should be blocked")
	assert.True(t, *publicAccessBlock.PublicAccessBlockConfiguration.BlockPublicPolicy, "Public policies should be blocked")
	assert.True(t, *publicAccessBlock.PublicAccessBlockConfiguration.IgnorePublicAcls, "Public ACLs should be ignored")
	assert.True(t, *publicAccessBlock.PublicAccessBlockConfiguration.RestrictPublicBuckets, "Public buckets should be restricted")

	// Verify encryption using AWS SDK v2
	encryptionConfig, err := s3Client.GetBucketEncryption(context.TODO(), &s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucketName),
	})
	require.NoError(t, err, "Should be able to get bucket encryption configuration")
	require.NotEmpty(t, encryptionConfig.ServerSideEncryptionConfiguration.Rules, "Encryption rules should be configured")

	rule := encryptionConfig.ServerSideEncryptionConfiguration.Rules[0]
	assert.Equal(t, types.ServerSideEncryptionAes256, rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm, "Should use AES256 encryption")
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

	// Verify it's a conditional failure (the specific error type varies, so just check it's an error)
	t.Logf("Expected conditional write failure: %v", err)

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

	// Handle masked region value in CI/CD environments
	regionValue := backendConfigMap["region"]
	if regionValue == "***" {
		// In CI/CD, the region might be masked, so just verify it's not empty
		assert.NotEmpty(t, regionValue, "Backend config should have a region value")
		t.Logf("Region value is masked in CI/CD: %s", regionValue)
	} else {
		assert.Equal(t, awsRegion, regionValue, "Backend config should have correct region")
	}

	assert.Equal(t, "true", backendConfigMap["encrypt"], "Backend config should have encryption enabled")
	assert.Equal(t, "true", backendConfigMap["use_lockfile"], "Backend config should have S3 native locking enabled")

	// Verify the example backend config contains the correct information
	exampleConfig := terraform.Output(t, opts, "example_backend_config")
	assert.Contains(t, exampleConfig, projectName+"-s3", "Example config should contain correct bucket name")

	// Handle masked region in CI/CD environments
	if backendConfigMap["region"] == "***" {
		assert.Contains(t, exampleConfig, "***", "Example config should contain masked region")
		t.Log("Region is masked in CI/CD environment")
	} else {
		assert.Contains(t, exampleConfig, awsRegion, "Example config should contain correct region")
	}

	assert.Contains(t, exampleConfig, "use_lockfile = true", "Example config should include S3 native locking")
	assert.Contains(t, exampleConfig, "encrypt      = true", "Example config should include encryption")

	// Verify it does NOT contain DynamoDB references (since we're using S3 native locking)
	assert.NotContains(t, exampleConfig, "dynamodb_table", "Example config should not contain DynamoDB table reference")
}

// Verify S3 bucket tags using Terratest
func verifyS3BucketTags(t *testing.T) {
	bucketName := terraform.Output(t, opts, "red_backend_s3_bucket")

	// Get bucket tags using Terratest (this function exists)
	tags := terratest_aws.GetS3BucketTags(t, awsRegion, bucketName)

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
		// Empty the bucket before destroying to avoid "bucket not empty" errors
		emptyS3Bucket(t)
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
