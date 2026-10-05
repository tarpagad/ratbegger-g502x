package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// backupVersion is bumped when the on-disk backup format changes incompatibly.
const backupVersion = 1

// BackupResolution captures one DPI slot. Slots are referenced by index, not by
// object path, because paths change when the device is re-plugged.
type BackupResolution struct {
	Index   uint32 `json:"index"`
	DPI     uint32 `json:"dpi"`
	Active  bool   `json:"active"`
	Default bool   `json:"default"`
}

// BackupButton captures one button mapping.
type BackupButton struct {
	Index      uint32       `json:"index"`
	ActionType uint32       `json:"actionType"`
	Value      uint32       `json:"value"`
	Macro      []MacroEvent `json:"macro,omitempty"`
}

// BackupProfile captures one profile.
type BackupProfile struct {
	Index       uint32             `json:"index"`
	Name        string             `json:"name"`
	Enabled     bool               `json:"enabled"`
	Active      bool               `json:"active"`
	ReportRate  uint32             `json:"reportRate"`
	Resolutions []BackupResolution `json:"resolutions"`
	Buttons     []BackupButton     `json:"buttons"`
}

// Backup is a complete snapshot of a device's onboard configuration.
type Backup struct {
	Version   int             `json:"version"`
	CreatedAt time.Time       `json:"createdAt"`
	Device    string          `json:"device"`
	Model     string          `json:"model"`
	Profiles  []BackupProfile `json:"profiles"`
}

// BackupInfo describes a stored backup for the UI.
type BackupInfo struct {
	Path      string    `json:"path"`
	File      string    `json:"file"`
	Device    string    `json:"device"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"createdAt"`
	Profiles  int       `json:"profiles"`
}

// SaveBackup snapshots the connected device (matched by model, or the first
// device when model is empty) and writes it to the backup directory.
func (a *App) SaveBackup(model string) (BackupInfo, error) {
	client, err := a.ratbagd()
	if err != nil {
		return BackupInfo{}, err
	}

	devices, err := client.Devices()
	if err != nil {
		return BackupInfo{}, err
	}
	device := findDevice(devices, model)
	if device == nil {
		return BackupInfo{}, fmt.Errorf("no connected device matched %q", model)
	}

	dir, err := backupDir()
	if err != nil {
		return BackupInfo{}, err
	}

	backup := backupFromDevice(*device)
	name := fmt.Sprintf("%s-%s.json", slug(device.Name), backup.CreatedAt.Format("2006-01-02T15-04-05"))
	path := filepath.Join(dir, name)

	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return BackupInfo{}, err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return BackupInfo{}, err
	}

	return infoFromBackup(path, backup), nil
}

// ListBackups returns the stored backups, newest first.
func (a *App) ListBackups() ([]BackupInfo, error) {
	dir, err := backupDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	infos := make([]BackupInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		backup, err := readBackup(path)
		if err != nil {
			continue
		}
		infos = append(infos, infoFromBackup(path, backup))
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].CreatedAt.After(infos[j].CreatedAt)
	})
	return infos, nil
}

// RestoreBackup applies a stored backup to the matching connected device.
// Before restoring it writes a fresh backup of the current state, so a restore
// that goes wrong can itself be undone.
func (a *App) RestoreBackup(path string) (BackupInfo, error) {
	if _, err := a.ratbagd(); err != nil {
		return BackupInfo{}, err
	}

	backup, err := readBackup(path)
	if err != nil {
		return BackupInfo{}, err
	}

	// Safety net: snapshot the current state first.
	pre, err := a.SaveBackup(backup.Model)
	if err != nil {
		return BackupInfo{}, fmt.Errorf("could not back up the current state first: %w", err)
	}

	client, err := a.ratbagd()
	if err != nil {
		return BackupInfo{}, err
	}
	devices, err := client.Devices()
	if err != nil {
		return BackupInfo{}, err
	}
	device := findDevice(devices, backup.Model)
	if device == nil {
		return BackupInfo{}, fmt.Errorf("no connected device matched %q", backup.Model)
	}

	restoreErr := client.RestoreBackup(*device, backup)
	return pre, restoreErr
}

// DeleteBackup removes a stored backup (only inside the backup directory).
func (a *App) DeleteBackup(path string) error {
	safe, err := safeBackupPath(path)
	if err != nil {
		return err
	}
	return os.Remove(safe)
}

// RestoreBackup writes a backup's values back onto a live device, then commits
// once. Errors are collected so one rejected value does not abort the rest.
func (r *Ratbagd) RestoreBackup(device Device, backup Backup) error {
	if backup.Version != backupVersion {
		return fmt.Errorf("unsupported backup version %d (this app writes %d)", backup.Version, backupVersion)
	}

	profiles := make(map[uint32]Profile, len(device.Profiles))
	for _, profile := range device.Profiles {
		profiles[profile.Index] = profile
	}

	var firstErr error
	record := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	for _, backupProfile := range backup.Profiles {
		profile, ok := profiles[backupProfile.Index]
		if !ok {
			continue
		}

		// Enable before writing bindings: disabled profiles may not hold them.
		record(r.SetProfileEnabled(profile.Path, backupProfile.Enabled))

		resolutions := make(map[uint32]Resolution, len(profile.Resolutions))
		for _, resolution := range profile.Resolutions {
			resolutions[resolution.Index] = resolution
		}
		for _, backupResolution := range backupProfile.Resolutions {
			resolution, ok := resolutions[backupResolution.Index]
			if !ok || !dpiAllowed(backupResolution.DPI, resolution.DPIs) {
				continue
			}
			record(r.SetResolutionDPI(resolution.Path, backupResolution.DPI))
		}

		buttons := make(map[uint32]Button, len(profile.Buttons))
		for _, button := range profile.Buttons {
			buttons[button.Index] = button
		}
		for _, backupButton := range backupProfile.Buttons {
			button, ok := buttons[backupButton.Index]
			if !ok {
				continue
			}
			switch backupButton.ActionType {
			case actionNone:
				record(r.DisableButton(button.Path))
			case actionMacro:
				record(r.SetButtonMacro(button.Path, backupButton.Macro))
			default:
				record(r.SetButtonMapping(button.Path, backupButton.ActionType, backupButton.Value))
			}
		}

		if backupProfile.ReportRate != 0 {
			record(r.SetReportRate(profile.Path, backupProfile.ReportRate))
		}
		if backupProfile.Name != "" {
			record(r.SetProfileName(profile.Path, backupProfile.Name))
		}

		// Default first, then active, so the active slot is not overwritten.
		for _, backupResolution := range backupProfile.Resolutions {
			if backupResolution.Default {
				if resolution, ok := resolutions[backupResolution.Index]; ok {
					record(r.SetResolutionDefault(resolution.Path))
				}
			}
		}
		for _, backupResolution := range backupProfile.Resolutions {
			if backupResolution.Active {
				if resolution, ok := resolutions[backupResolution.Index]; ok {
					record(r.SetResolutionActive(resolution.Path))
				}
			}
		}
	}

	for _, backupProfile := range backup.Profiles {
		if backupProfile.Active {
			if profile, ok := profiles[backupProfile.Index]; ok {
				record(r.SetProfileActive(profile.Path))
			}
		}
	}

	record(r.Commit(device.Path))
	return firstErr
}

func backupFromDevice(device Device) Backup {
	backup := Backup{
		Version:   backupVersion,
		CreatedAt: time.Now(),
		Device:    device.Name,
		Model:     device.Model,
	}

	for _, profile := range device.Profiles {
		backupProfile := BackupProfile{
			Index:      profile.Index,
			Name:       profile.Name,
			Enabled:    profile.Enabled,
			Active:     profile.Active,
			ReportRate: profile.ReportRate,
		}
		for _, resolution := range profile.Resolutions {
			backupProfile.Resolutions = append(backupProfile.Resolutions, BackupResolution{
				Index:   resolution.Index,
				DPI:     resolution.DPI,
				Active:  resolution.Active,
				Default: resolution.Default,
			})
		}
		for _, button := range profile.Buttons {
			backupProfile.Buttons = append(backupProfile.Buttons, BackupButton{
				Index:      button.Index,
				ActionType: button.ActionType,
				Value:      button.Value,
				Macro:      button.Macro,
			})
		}
		backup.Profiles = append(backup.Profiles, backupProfile)
	}

	return backup
}

func dpiAllowed(dpi uint32, allowed []uint32) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, candidate := range allowed {
		if candidate == dpi {
			return true
		}
	}
	return false
}

func findDevice(devices []Device, model string) *Device {
	for i := range devices {
		if model == "" || devices[i].Model == model {
			return &devices[i]
		}
	}
	return nil
}

func backupDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate config directory: %w", err)
	}
	dir := filepath.Join(base, "ratbegger-g502x", "backups")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func readBackup(path string) (Backup, error) {
	safe, err := safeBackupPath(path)
	if err != nil {
		return Backup{}, err
	}

	data, err := os.ReadFile(safe)
	if err != nil {
		return Backup{}, err
	}

	var backup Backup
	if err := json.Unmarshal(data, &backup); err != nil {
		return Backup{}, fmt.Errorf("read backup %s: %w", filepath.Base(safe), err)
	}
	return backup, nil
}

// safeBackupPath rejects paths outside the backup directory.
func safeBackupPath(path string) (string, error) {
	dir, err := backupDir()
	if err != nil {
		return "", err
	}

	cleaned := filepath.Clean(path)
	relative, err := filepath.Rel(dir, cleaned)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("refusing to touch a file outside the backup directory")
	}
	return cleaned, nil
}

func infoFromBackup(path string, backup Backup) BackupInfo {
	return BackupInfo{
		Path:      path,
		File:      filepath.Base(path),
		Device:    backup.Device,
		Model:     backup.Model,
		CreatedAt: backup.CreatedAt,
		Profiles:  len(backup.Profiles),
	}
}

func slug(value string) string {
	mapped := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		default:
			return '-'
		}
	}, value)
	mapped = strings.Trim(mapped, "-")
	if mapped == "" {
		return "device"
	}
	return mapped
}
