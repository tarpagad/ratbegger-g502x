package main

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	busService  = "org.freedesktop.ratbag1"
	managerPath = dbus.ObjectPath("/org/freedesktop/ratbag1")

	ifaceDevice  = "org.freedesktop.ratbag1.Device"
	ifaceManager = "org.freedesktop.ratbag1.Manager"
	ifaceProfile = "org.freedesktop.ratbag1.Profile"
	ifaceButton  = "org.freedesktop.ratbag1.Button"
	ifaceRes     = "org.freedesktop.ratbag1.Resolution"
)

// expectedAPIVersion is the ratbagd D-Bus API version this client understands.
// The API has no compatibility guarantees: the version must match exactly.
const expectedAPIVersion = 2

// macroEvent mirrors the (uu) tuple libratbag uses for macro steps.
type macroEvent struct {
	Type    uint32
	Keycode uint32
}

// Ratbagd is a thin client over the ratbagd system D-Bus service.
type Ratbagd struct {
	conn *dbus.Conn
}

// Connect opens the system bus. ratbagd itself is contacted lazily.
func Connect() (*Ratbagd, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("connect to system bus: %w", err)
	}
	return &Ratbagd{conn: conn}, nil
}

// Close releases the bus connection.
func (r *Ratbagd) Close() {
	if r != nil && r.conn != nil {
		r.conn.Close()
	}
}

// APIVersion returns ratbagd's D-Bus API version.
func (r *Ratbagd) APIVersion() (int32, error) {
	var v int32
	if err := r.store(managerPath, ifaceManager, "APIVersion", &v); err != nil {
		return 0, fmt.Errorf("ratbagd is not reachable: %w", err)
	}
	return v, nil
}

// WatchResync calls fn whenever a device emits the Resync signal, which
// ratbagd sends when a commit failed and clients must reload their state.
func (r *Ratbagd) WatchResync(fn func(devicePath string)) error {
	if err := r.conn.AddMatchSignal(
		dbus.WithMatchInterface(ifaceDevice),
		dbus.WithMatchMember("Resync"),
	); err != nil {
		return err
	}

	ch := make(chan *dbus.Signal, 8)
	r.conn.Signal(ch)
	go func() {
		for sig := range ch {
			if sig.Name == ifaceDevice+".Resync" {
				fn(string(sig.Path))
			}
		}
	}()
	return nil
}

// Devices walks every device and its profile/resolution/button tree.
func (r *Ratbagd) Devices() ([]Device, error) {
	var paths []dbus.ObjectPath
	if err := r.store(managerPath, ifaceManager, "Devices", &paths); err != nil {
		return nil, err
	}

	devices := make([]Device, 0, len(paths))
	for _, path := range paths {
		device, err := r.device(path)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	return devices, nil
}

// Commit writes any pending property changes to the device. It is
// asynchronous: errors surface later as a Resync signal.
func (r *Ratbagd) Commit(devicePath string) error {
	return r.call(devicePath, ifaceDevice+".Commit")
}

func (r *Ratbagd) device(path dbus.ObjectPath) (Device, error) {
	device := Device{Path: string(path)}
	if err := r.store(path, ifaceDevice, "Name", &device.Name); err != nil {
		return device, err
	}
	if err := r.store(path, ifaceDevice, "Model", &device.Model); err != nil {
		return device, err
	}

	profiles, err := r.objectPaths(path, ifaceDevice, "Profiles")
	if err != nil {
		return device, err
	}
	for _, profilePath := range profiles {
		profile, err := r.profile(profilePath)
		if err != nil {
			return device, err
		}
		device.Profiles = append(device.Profiles, profile)
	}
	return device, nil
}

func (r *Ratbagd) profile(path dbus.ObjectPath) (Profile, error) {
	profile := Profile{Path: string(path)}
	if err := r.store(path, ifaceProfile, "Index", &profile.Index); err != nil {
		return profile, err
	}
	_ = r.store(path, ifaceProfile, "Name", &profile.Name)
	if err := r.store(path, ifaceProfile, "IsActive", &profile.Active); err != nil {
		return profile, err
	}
	if err := r.store(path, ifaceProfile, "ReportRate", &profile.ReportRate); err != nil {
		return profile, err
	}
	_ = r.store(path, ifaceProfile, "ReportRates", &profile.ReportRates)

	enabled, err := r.profileEnabled(path)
	if err != nil {
		return profile, err
	}
	profile.Enabled = enabled

	resolutions, err := r.objectPaths(path, ifaceProfile, "Resolutions")
	if err != nil {
		return profile, err
	}
	for _, resPath := range resolutions {
		resolution, err := r.resolution(resPath)
		if err != nil {
			return profile, err
		}
		profile.Resolutions = append(profile.Resolutions, resolution)
	}

	buttons, err := r.objectPaths(path, ifaceProfile, "Buttons")
	if err != nil {
		return profile, err
	}
	for _, buttonPath := range buttons {
		button, err := r.button(buttonPath)
		if err != nil {
			return profile, err
		}
		profile.Buttons = append(profile.Buttons, button)
	}
	return profile, nil
}

func (r *Ratbagd) resolution(path dbus.ObjectPath) (Resolution, error) {
	resolution := Resolution{Path: string(path)}
	if err := r.store(path, ifaceRes, "Index", &resolution.Index); err != nil {
		return resolution, err
	}
	if err := r.store(path, ifaceRes, "IsActive", &resolution.Active); err != nil {
		return resolution, err
	}
	if err := r.store(path, ifaceRes, "IsDefault", &resolution.Default); err != nil {
		return resolution, err
	}
	_ = r.store(path, ifaceRes, "Resolutions", &resolution.DPIs)

	dpi, err := r.resolutionDPI(path)
	if err != nil {
		return resolution, err
	}
	resolution.DPI = dpi
	return resolution, nil
}

func (r *Ratbagd) button(path dbus.ObjectPath) (Button, error) {
	button := Button{Path: string(path)}
	if err := r.store(path, ifaceButton, "Index", &button.Index); err != nil {
		return button, err
	}
	_ = r.store(path, ifaceButton, "ActionTypes", &button.ActionTypes)

	var mapping struct {
		ActionType uint32
		Payload    dbus.Variant
	}
	if err := r.store(path, ifaceButton, "Mapping", &mapping); err != nil {
		return button, err
	}
	button.ActionType = mapping.ActionType

	switch payload := mapping.Payload.Value().(type) {
	case uint32:
		button.Value = payload
	case [][]interface{}:
		button.Macro = macrosFromPayload(payload)
	}
	return button, nil
}

// macrosFromPayload converts godbus' generic a(uu) representation into
// MacroEvent values. Each element is a []interface{}{type, keycode}.
func macrosFromPayload(payload [][]interface{}) []MacroEvent {
	events := make([]MacroEvent, 0, len(payload))
	for _, pair := range payload {
		if len(pair) != 2 {
			continue
		}
		eventType, _ := pair[0].(uint32)
		keycode, _ := pair[1].(uint32)
		events = append(events, MacroEvent{Type: eventType, Keycode: keycode})
	}
	return events
}

// resolutionDPI reads the Resolution property, which is a variant wrapper.
func (r *Ratbagd) resolutionDPI(path dbus.ObjectPath) (uint32, error) {
	var value dbus.Variant
	if err := r.store(path, ifaceRes, "Resolution", &value); err != nil {
		return 0, err
	}

	// The property is itself a variant, so godbus often hands back a
	// variant nested inside the one we asked for.
	if inner, ok := value.Value().(dbus.Variant); ok {
		value = inner
	}

	switch dpi := value.Value().(type) {
	case uint32:
		return dpi, nil
	case []interface{}: // (uu): separate x/y resolution
		if len(dpi) == 2 {
			if x, ok := dpi[0].(uint32); ok {
				return x, nil
			}
		}
	}
	return 0, fmt.Errorf("unexpected resolution value %T", value.Value())
}

// SetResolutionDPI writes a DPI value to a resolution slot.
func (r *Ratbagd) SetResolutionDPI(path string, dpi uint32) error {
	return r.setProperty(path, ifaceRes, "Resolution", dbus.MakeVariant(dpi))
}

// SetResolutionActive makes a resolution slot the active one.
func (r *Ratbagd) SetResolutionActive(path string) error {
	return r.call(path, ifaceRes+".SetActive")
}

// SetResolutionDefault makes a resolution slot the default one.
func (r *Ratbagd) SetResolutionDefault(path string) error {
	return r.call(path, ifaceRes+".SetDefault")
}

// SetReportRate sets a profile's report rate in Hz.
func (r *Ratbagd) SetReportRate(path string, rate uint32) error {
	return r.setProperty(path, ifaceProfile, "ReportRate", rate)
}

// SetProfileActive makes a profile the active one.
func (r *Ratbagd) SetProfileActive(path string) error {
	return r.call(path, ifaceProfile+".SetActive")
}

// SetProfileName renames a profile. Devices that cannot store names expose an
// empty Name property and ignore writes.
func (r *Ratbagd) SetProfileName(path, name string) error {
	return r.setProperty(path, ifaceProfile, "Name", name)
}

// SetProfileEnabled enables or disables a profile. Depending on the libratbag
// version this is exposed either as "Disabled" (0.18) or "Enabled" (newer).
func (r *Ratbagd) SetProfileEnabled(path string, enabled bool) error {
	prop, inverted, err := r.enabledProperty(path)
	if err != nil {
		return err
	}
	value := enabled
	if inverted {
		value = !enabled
	}
	return r.setProperty(path, ifaceProfile, prop, value)
}

// SetButtonMapping assigns a None/Button/Special/Key action to a button.
func (r *Ratbagd) SetButtonMapping(path string, actionType, value uint32) error {
	if actionType == actionNone {
		return r.call(path, ifaceButton+".Disable")
	}

	mapping := struct {
		ActionType uint32
		Payload    dbus.Variant
	}{actionType, dbus.MakeVariant(value)}
	return r.setProperty(path, ifaceButton, "Mapping", mapping)
}

// SetButtonMacro assigns a recorded key sequence to a button.
func (r *Ratbagd) SetButtonMacro(path string, events []MacroEvent) error {
	pairs := make([]macroEvent, len(events))
	for i, event := range events {
		pairs[i] = macroEvent{Type: event.Type, Keycode: event.Keycode}
	}

	mapping := struct {
		ActionType uint32
		Payload    dbus.Variant
	}{actionMacro, dbus.MakeVariant(pairs)}
	return r.setProperty(path, ifaceButton, "Mapping", mapping)
}

// DisableButton removes any mapping from a button.
func (r *Ratbagd) DisableButton(path string) error {
	return r.call(path, ifaceButton+".Disable")
}

// enabledProperty reports which property reflects a profile's enabled state
// and whether reading it must be inverted.
func (r *Ratbagd) enabledProperty(path string) (string, bool, error) {
	var disabled bool
	if err := r.store(dbus.ObjectPath(path), ifaceProfile, "Disabled", &disabled); err == nil {
		return "Disabled", true, nil
	}
	var enabled bool
	if err := r.store(dbus.ObjectPath(path), ifaceProfile, "Enabled", &enabled); err == nil {
		return "Enabled", false, nil
	}
	return "", false, fmt.Errorf("profile exposes neither a Disabled nor an Enabled property")
}

func (r *Ratbagd) profileEnabled(path dbus.ObjectPath) (bool, error) {
	prop, inverted, err := r.enabledProperty(string(path))
	if err != nil {
		return false, err
	}
	var value bool
	if err := r.store(path, ifaceProfile, prop, &value); err != nil {
		return false, err
	}
	if inverted {
		return !value, nil
	}
	return value, nil
}

func (r *Ratbagd) objectPaths(path dbus.ObjectPath, iface, prop string) ([]dbus.ObjectPath, error) {
	var paths []dbus.ObjectPath
	if err := r.store(path, iface, prop, &paths); err != nil {
		return nil, err
	}
	return paths, nil
}

func (r *Ratbagd) store(path dbus.ObjectPath, iface, prop string, dest interface{}) error {
	return r.conn.Object(busService, path).StoreProperty(iface+"."+prop, dest)
}

func (r *Ratbagd) setProperty(path, iface, prop string, value interface{}) error {
	return r.conn.Object(busService, dbus.ObjectPath(path)).
		SetProperty(iface+"."+prop, value)
}

func (r *Ratbagd) call(path, method string) error {
	return r.conn.Object(busService, dbus.ObjectPath(path)).Call(method, 0).Err
}
