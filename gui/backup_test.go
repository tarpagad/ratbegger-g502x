package main

import (
	"encoding/json"
	"testing"
)

func TestBackupFromDevice(t *testing.T) {
	device := Device{
		Path:  "/org/freedesktop/ratbag1/device/hidraw0",
		Name:  "Logitech G502 X",
		Model: "usb:046d:c099:0",
		Profiles: []Profile{{
			Path:        "/org/freedesktop/ratbag1/profile/hidraw0/p0",
			Index:       0,
			Enabled:     true,
			Active:      true,
			ReportRate:  1000,
			ReportRates: []uint32{125, 250, 500, 1000},
			Resolutions: []Resolution{{
				Path:  "/org/freedesktop/ratbag1/resolution/hidraw0/p0/r0",
				Index: 0, DPI: 800, Active: true, Default: true,
				DPIs: []uint32{800, 1600},
			}},
			Buttons: []Button{{
				Path:  "/org/freedesktop/ratbag1/button/hidraw0/p0/b9",
				Index: 9, ActionType: actionMacro,
				Macro: []MacroEvent{{Type: macroPress, Keycode: 29}, {Type: macroRelease, Keycode: 29}},
			}},
		}},
	}

	backup := backupFromDevice(device)

	if backup.Version != backupVersion {
		t.Errorf("version = %d, want %d", backup.Version, backupVersion)
	}
	if backup.Device != "Logitech G502 X" || backup.Model != "usb:046d:c099:0" {
		t.Errorf("unexpected device identity: %+v", backup)
	}
	if len(backup.Profiles) != 1 {
		t.Fatalf("got %d profiles, want 1", len(backup.Profiles))
	}

	profile := backup.Profiles[0]
	if profile.Index != 0 || !profile.Active || !profile.Enabled || profile.ReportRate != 1000 {
		t.Errorf("unexpected profile: %+v", profile)
	}
	if len(profile.Resolutions) != 1 || profile.Resolutions[0].DPI != 800 || !profile.Resolutions[0].Default {
		t.Errorf("unexpected resolutions: %+v", profile.Resolutions)
	}
	if len(profile.Buttons) != 1 || profile.Buttons[0].ActionType != actionMacro {
		t.Fatalf("unexpected buttons: %+v", profile.Buttons)
	}
	if len(profile.Buttons[0].Macro) != 2 || profile.Buttons[0].Macro[0].Keycode != 29 {
		t.Errorf("unexpected macro: %+v", profile.Buttons[0].Macro)
	}
}

func TestBackupJSONRoundTrip(t *testing.T) {
	original := Backup{
		Version: backupVersion,
		Device:  "Logitech G502 X",
		Model:   "usb:046d:c099:0",
		Profiles: []BackupProfile{{
			Index:      1,
			Enabled:    true,
			ReportRate: 500,
			Buttons:    []BackupButton{{Index: 4, ActionType: actionSpecial, Value: 0x4000000b}},
		}},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Backup
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Model != original.Model || len(decoded.Profiles) != 1 {
		t.Fatalf("round trip changed the backup: %+v", decoded)
	}
	button := decoded.Profiles[0].Buttons[0]
	if button.ActionType != actionSpecial || button.Value != 0x4000000b {
		t.Errorf("unexpected button after round trip: %+v", button)
	}
}

func TestDPIAllowed(t *testing.T) {
	allowed := []uint32{800, 1600, 2400}

	if !dpiAllowed(1600, allowed) {
		t.Error("1600 should be allowed")
	}
	if dpiAllowed(1234, allowed) {
		t.Error("1234 should not be allowed")
	}
	if !dpiAllowed(99999, nil) {
		t.Error("anything should be allowed when the list is empty")
	}
}

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Logitech G502 X": "Logitech-G502-X",
		"a/b:c":           "a-b-c",
		"":                "device",
		"---":             "device",
	}
	for input, want := range cases {
		if got := slug(input); got != want {
			t.Errorf("slug(%q) = %q, want %q", input, got, want)
		}
	}
}
