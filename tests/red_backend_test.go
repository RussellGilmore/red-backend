package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
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

func destroyTerraform(t *testing.T) {
	terraform.Destroy(t, opts)
}

func deployTerraform(t *testing.T) {
	_, err := terraform.InitAndApplyE(t, opts)
	if err != nil {
		terraform.Apply(t, opts)
	}
}

func verifyRedBackendNames(t *testing.T) {
	bucketName := terraform.Output(t, opts, "red_backend_s3_bucket")
	tableName := terraform.Output(t, opts, "red_backend_ddb_table")

	expectedBucketName := projectName + "-s3"
	expectedTableName := projectName + "-tf-lock-status"

	assert.Equal(t, expectedTableName, tableName)

	assert.Equal(t, expectedBucketName, bucketName)
}

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
}
