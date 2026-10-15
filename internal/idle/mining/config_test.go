package mining

import "testing"

func TestDisabledConfigurationNeedsNoCredentials(t *testing.T) {
	if err := (Config{}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEnabledProfilesRequireOwnerConfiguration(t *testing.T) {
	for _, kind := range []string{"monero", "verus"} {
		c := Config{Enabled: true, Kind: kind, CPUPercent: 50}
		if err := c.Validate(); err == nil {
			t.Fatalf("%s accepted missing owner configuration", kind)
		}
	}
}

func TestConfiguredProfilesValidate(t *testing.T) {
	for _, kind := range []string{"monero", "verus"} {
		c := Config{
			Enabled: true, Kind: kind, Executable: "/external/worker",
			Pool: "owner-configured", Wallet: "owner-configured", CPUPercent: 50,
		}
		if err := c.Validate(); err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
	}
}
