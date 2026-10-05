package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestLiveRestoreRoundTrip is the only test that writes to the device. It takes
// a backup of the current settings, restores that same backup, and asserts the
// device reads back identically. Because the backup equals the current state,
// the restore is a no-op in effect — but it exercises the whole write path.
//
// It is gated behind its own env var so it never runs with the ordinary live
// tests, and it writes its safety backups to the real config directory so the
// current settings are recoverable:
//
//	RATBAGD_LIVE_WRITE=1 go test -run TestLiveRestoreRoundTrip -v .
func TestLiveRestoreRoundTrip(t *testing.T) {
	if os.Getenv("RATBAGD_LIVE_WRITE") == "" {
		t.Skip("set RATBAGD_LIVE_WRITE=1 to run the device-writing restore test")
	}

	client, err := Connect()
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer client.Close()

	before, err := client.Devices()
	if err != nil {
		t.Fatalf("read before: %v", err)
	}
	beforeSignature := backupSignature(before)

	app := &App{client: client}
	safety, err := app.SaveBackup("")
	if err != nil {
		t.Fatalf("safety backup: %v", err)
	}
	t.Logf("safety backup written to %s", safety.Path)

	if _, err := app.RestoreBackup(safety.Path); err != nil {
		t.Fatalf("restore failed: %v (safety backup at %s)", err, safety.Path)
	}

	// Commit is asynchronous; give ratbagd a moment to write.
	time.Sleep(2 * time.Second)

	after, err := client.Devices()
	if err != nil {
		t.Fatalf("read after: %v", err)
	}
	afterSignature := backupSignature(after)

	if beforeSignature != afterSignature {
		t.Fatalf("device state changed after restoring an identical backup\nbefore: %s\nafter:  %s",
			beforeSignature, afterSignature)
	}

	t.Logf("device state identical after restore (%d device(s))", len(after))
}

// TestLiveWriteAndRevert proves the write path actually takes effect: it changes
// one DPI value on an inactive profile, confirms the device reports the new
// value, then restores the recovery backup and confirms the original is back.
// Gated behind RATBAGD_LIVE_WRITE=1 like the round-trip test.
func TestLiveWriteAndRevert(t *testing.T) {
	if os.Getenv("RATBAGD_LIVE_WRITE") == "" {
		t.Skip("set RATBAGD_LIVE_WRITE=1 to run the device-writing tests")
	}

	client, err := Connect()
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer client.Close()

	app := &App{client: client}
	recovery, err := app.SaveBackup("")
	if err != nil {
		t.Fatalf("recovery backup: %v", err)
	}
	t.Logf("recovery backup written to %s", recovery.Path)

	devices, err := client.Devices()
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	// Use an inactive profile so the change does not affect what the user feels.
	profile, ok := pickInactiveProfile(devices[0])
	if !ok {
		t.Skip("no inactive profile with a resolution to edit")
	}
	original := profile.Resolutions[0]
	target := differentDPI(original.DPI, original.DPIs)
	if target == original.DPI {
		t.Skip("no alternative DPI available")
	}

	if err := client.SetResolutionDPI(original.Path, target); err != nil {
		t.Fatalf("set dpi: %v", err)
	}
	if err := client.Commit(devices[0].Path); err != nil {
		t.Fatalf("commit: %v", err)
	}
	time.Sleep(2 * time.Second)

	if got := readDPI(t, client, profile.Index, original.Index); got != target {
		t.Fatalf("after write, DPI = %d, want %d", got, target)
	}
	t.Logf("write applied: profile %d resolution %d: %d -> %d",
		profile.Index, original.Index, original.DPI, target)

	if _, err := app.RestoreBackup(recovery.Path); err != nil {
		t.Fatalf("restore: %v", err)
	}
	time.Sleep(2 * time.Second)

	if got := readDPI(t, client, profile.Index, original.Index); got != original.DPI {
		t.Fatalf("after restore, DPI = %d, want %d", got, original.DPI)
	}
	t.Logf("restore reverted: profile %d resolution %d back to %d",
		profile.Index, original.Index, original.DPI)
}

func pickInactiveProfile(device Device) (Profile, bool) {
	for _, profile := range device.Profiles {
		if !profile.Active && len(profile.Resolutions) > 0 {
			return profile, true
		}
	}
	return Profile{}, false
}

// differentDPI returns the allowed value next to current, so the change is as
// small as possible.
func differentDPI(current uint32, allowed []uint32) uint32 {
	for i, dpi := range allowed {
		if dpi != current {
			continue
		}
		if i+1 < len(allowed) {
			return allowed[i+1]
		}
		if i > 0 {
			return allowed[i-1]
		}
	}
	for _, dpi := range allowed {
		if dpi != current {
			return dpi
		}
	}
	return current
}

func readDPI(t *testing.T, client *Ratbagd, profileIndex, resolutionIndex uint32) uint32 {
	t.Helper()
	devices, err := client.Devices()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	for _, profile := range devices[0].Profiles {
		if profile.Index != profileIndex {
			continue
		}
		for _, resolution := range profile.Resolutions {
			if resolution.Index == resolutionIndex {
				return resolution.DPI
			}
		}
	}
	t.Fatalf("profile %d resolution %d not found", profileIndex, resolutionIndex)
	return 0
}

// backupSignature is a stable, comparable rendering of the devices' settings
// (index-keyed, timestamps removed).
func backupSignature(devices []Device) string {
	backups := make([]Backup, 0, len(devices))
	for _, device := range devices {
		backup := backupFromDevice(device)
		backup.CreatedAt = time.Time{}
		backups = append(backups, backup)
	}
	data, err := json.Marshal(backups)
	if err != nil {
		return "marshal error: " + err.Error()
	}
	return string(data)
}
