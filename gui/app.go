package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// App is the Wails application backend. It shells out to ratbagctl, the CLI
// provided by libratbag (https://github.com/libratbag/libratbag). Changes are
// written to the mouse's onboard memory, so no daemon needs to keep running
// after a setting is applied.
//
// This is a minimal starting point: every call maps to one ratbagctl command
// and returns its raw output for the UI to display.
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Device is a libratbag-supported device, as reported by "ratbagctl list".
type Device struct {
	Codename string `json:"codename"`
	Name     string `json:"name"`
}

// RatbagctlAvailable reports whether ratbagctl is on PATH.
func (a *App) RatbagctlAvailable() bool {
	_, err := exec.LookPath("ratbagctl")
	return err == nil
}

// ListDevices runs "ratbagctl list" and parses lines of the form
// "<codename>: <name>".
func (a *App) ListDevices() ([]Device, error) {
	out, err := runRatbagctl("ratbagctl", "list")
	if err != nil {
		return nil, err
	}

	var devices []Device
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		codename, name, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		devices = append(devices, Device{
			Codename: strings.TrimSpace(codename),
			Name:     strings.TrimSpace(name),
		})
	}
	return devices, nil
}

// GetDeviceInfo returns the raw output of "ratbagctl <device> info".
func (a *App) GetDeviceInfo(device string) (string, error) {
	return runRatbagctl("ratbagctl", device, "info")
}

// SetDPI writes a DPI value to a resolution slot of a profile.
func (a *App) SetDPI(device string, profile, slot, dpi int) (string, error) {
	return runRatbagctl("ratbagctl", device,
		"profile", fmt.Sprint(profile),
		"resolution", fmt.Sprint(slot),
		"dpi", "set", fmt.Sprint(dpi))
}

// SetReportRate sets the USB report rate (125|250|500|1000) of a profile.
func (a *App) SetReportRate(device string, profile, rate int) (string, error) {
	return runRatbagctl("ratbagctl", device,
		"profile", fmt.Sprint(profile),
		"rate", "set", fmt.Sprint(rate))
}

// SetActiveProfile makes the given profile (0-4) the active one.
func (a *App) SetActiveProfile(device string, profile int) (string, error) {
	return runRatbagctl("ratbagctl", device,
		"profile", "active", "set", fmt.Sprint(profile))
}

// runRatbagctl executes a command without a shell and returns its combined
// output. Arguments are passed verbatim, so a device name containing spaces is
// safe.
func runRatbagctl(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s: %w", strings.Join(append([]string{name}, args...), " "), err)
	}
	return string(out), nil
}
