package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails application backend. It talks to ratbagd over the system
// D-Bus instead of spawning ratbagctl, so every change goes through the same
// code path the CLI and Piper use.
type App struct {
	ctx    context.Context
	client *Ratbagd
	err    error
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. It connects to ratbagd and subscribes
// to device Resync signals, which are forwarded to the frontend.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	client, err := Connect()
	if err != nil {
		a.err = err
		return
	}
	a.client = client

	if err := client.WatchResync(func(devicePath string) {
		runtime.EventsEmit(ctx, "ratbagd:resync", devicePath)
	}); err != nil {
		a.err = err
	}
}

func (a *App) ratbagd() (*Ratbagd, error) {
	if a.err != nil {
		return nil, a.err
	}
	if a.client == nil {
		return nil, fmt.Errorf("ratbagd client is not initialised")
	}
	return a.client, nil
}

// Status reports whether ratbagd is reachable and which API version it speaks.
func (a *App) Status() Status {
	client, err := a.ratbagd()
	if err != nil {
		return Status{Error: err.Error()}
	}

	version, err := client.APIVersion()
	if err != nil {
		return Status{Error: err.Error()}
	}

	status := Status{Connected: true, APIVersion: version}
	if version != expectedAPIVersion {
		status.Connected = false
		status.Error = fmt.Sprintf(
			"unsupported ratbagd D-Bus API version %d (this app needs %d)",
			version, expectedAPIVersion,
		)
	}
	return status
}

// ListDevices returns every device with its full profile tree.
func (a *App) ListDevices() ([]Device, error) {
	client, err := a.ratbagd()
	if err != nil {
		return nil, err
	}
	return client.Devices()
}

// SetResolution writes a DPI value and commits it to the device.
func (a *App) SetResolution(devicePath, resolutionPath string, dpi uint32) error {
	client, err := a.ratbagd()
	if err != nil {
		return err
	}
	if err := client.SetResolutionDPI(resolutionPath, dpi); err != nil {
		return err
	}
	return client.Commit(devicePath)
}

// SetResolutionActive activates a resolution slot and commits.
func (a *App) SetResolutionActive(devicePath, resolutionPath string) error {
	client, err := a.ratbagd()
	if err != nil {
		return err
	}
	if err := client.SetResolutionActive(resolutionPath); err != nil {
		return err
	}
	return client.Commit(devicePath)
}

// SetResolutionDefault sets the default resolution slot and commits.
func (a *App) SetResolutionDefault(devicePath, resolutionPath string) error {
	client, err := a.ratbagd()
	if err != nil {
		return err
	}
	if err := client.SetResolutionDefault(resolutionPath); err != nil {
		return err
	}
	return client.Commit(devicePath)
}

// SetReportRate sets a profile's report rate and commits.
func (a *App) SetReportRate(devicePath, profilePath string, rate uint32) error {
	client, err := a.ratbagd()
	if err != nil {
		return err
	}
	if err := client.SetReportRate(profilePath, rate); err != nil {
		return err
	}
	return client.Commit(devicePath)
}

// SetProfileActive makes a profile active and commits.
func (a *App) SetProfileActive(devicePath, profilePath string) error {
	client, err := a.ratbagd()
	if err != nil {
		return err
	}
	if err := client.SetProfileActive(profilePath); err != nil {
		return err
	}
	return client.Commit(devicePath)
}

// SetProfileEnabled enables or disables a profile and commits.
func (a *App) SetProfileEnabled(devicePath, profilePath string, enabled bool) error {
	client, err := a.ratbagd()
	if err != nil {
		return err
	}
	if err := client.SetProfileEnabled(profilePath, enabled); err != nil {
		return err
	}
	return client.Commit(devicePath)
}

// SetProfileName renames a profile and commits.
func (a *App) SetProfileName(devicePath, profilePath, name string) error {
	client, err := a.ratbagd()
	if err != nil {
		return err
	}
	if err := client.SetProfileName(profilePath, name); err != nil {
		return err
	}
	return client.Commit(devicePath)
}

// SetButtonAction assigns a None/Button/Special/Key action and commits.
func (a *App) SetButtonAction(devicePath, buttonPath string, actionType, value uint32) error {
	client, err := a.ratbagd()
	if err != nil {
		return err
	}
	if err := client.SetButtonMapping(buttonPath, actionType, value); err != nil {
		return err
	}
	return client.Commit(devicePath)
}

// SetButtonMacro assigns a recorded key sequence and commits.
func (a *App) SetButtonMacro(devicePath, buttonPath string, events []MacroEvent) error {
	client, err := a.ratbagd()
	if err != nil {
		return err
	}
	if err := client.SetButtonMacro(buttonPath, events); err != nil {
		return err
	}
	return client.Commit(devicePath)
}

// DisableButton removes a button's mapping and commits.
func (a *App) DisableButton(devicePath, buttonPath string) error {
	client, err := a.ratbagd()
	if err != nil {
		return err
	}
	if err := client.DisableButton(buttonPath); err != nil {
		return err
	}
	return client.Commit(devicePath)
}
