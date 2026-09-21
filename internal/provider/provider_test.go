package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"rustrak": testAccProvider,
	}
}

// TestProvider validates the provider's schema is internally consistent.
// This alone won't catch the resourceApplication schema gap unless its
// CreateContext is exercised, but it's a fast first check.
func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("provider InternalValidate failed: %s", err)
	}
}

// testAccPreCheck ensures required env vars are set before any TF_ACC test
// runs. resource.Test itself skips the whole test if TF_ACC is unset, so
// these acceptance tests never run in a normal `go test ./...`.
func testAccPreCheck(t *testing.T) {
	if os.Getenv("RUSTRAK_TOKEN") == "" {
		t.Fatal("RUSTRAK_TOKEN must be set for acceptance tests")
	}
}

func TestAccrustrakSecret_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig("TEST_KEY", "test-value-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rustrak_secret.test", "key", "TEST_KEY"),
					resource.TestCheckResourceAttr("rustrak_secret.test", "value", "test-value-1"),
					resource.TestCheckResourceAttrSet("rustrak_secret.test", "id"),
					resource.TestCheckResourceAttrSet("rustrak_secret.test", "version"),
				),
			},
			{
				// Second step re-applies with a changed value to exercise the
				// update path (resourceSecretUpdate), not just create.
				Config: testAccSecretConfig("TEST_KEY", "test-value-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rustrak_secret.test", "value", "test-value-2"),
				),
			},
		},
	})
}

func testAccSecretConfig(key, value string) string {
	return fmt.Sprintf(`
resource "rustrak_secret" "test" {
  env    = "development"
  key    = %q
  value  = %q
}
`, key, value)
}
