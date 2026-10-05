package main

import (
	"os"
	"testing"
)

// TestLiveDevices walks a real ratbagd. It is opt-in so the default test run
// needs neither a daemon nor the hardware:
//
//	RATBAGD_LIVE=1 go test -run TestLiveDevices -v .
func TestLiveDevices(t *testing.T) {
	if os.Getenv("RATBAGD_LIVE") == "" {
		t.Skip("set RATBAGD_LIVE=1 to run against a real ratbagd")
	}

	client, err := Connect()
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer client.Close()

	version, err := client.APIVersion()
	if err != nil {
		t.Fatalf("api version: %v", err)
	}
	t.Logf("APIVersion = %d", version)

	devices, err := client.Devices()
	if err != nil {
		t.Fatalf("devices: %v", err)
	}
	if len(devices) == 0 {
		t.Fatal("no devices found")
	}

	for _, device := range devices {
		t.Logf("device %q (%s) at %s", device.Name, device.Model, device.Path)
		for _, profile := range device.Profiles {
			t.Logf("  profile %d name=%q active=%v enabled=%v rate=%d rates=%v",
				profile.Index, profile.Name, profile.Active, profile.Enabled,
				profile.ReportRate, profile.ReportRates)
			for _, resolution := range profile.Resolutions {
				t.Logf("    resolution %d dpi=%d active=%v default=%v choices=%d",
					resolution.Index, resolution.DPI, resolution.Active,
					resolution.Default, len(resolution.DPIs))
			}
			for _, button := range profile.Buttons {
				t.Logf("    button %d actionType=%d value=%d macro=%v types=%v",
					button.Index, button.ActionType, button.Value,
					button.Macro, button.ActionTypes)
			}
		}
	}
}
