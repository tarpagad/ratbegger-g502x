package main

// libratbag button action types (enum ratbag_button_action_type).
// Note: the published D-Bus docs list these in the wrong order — KEY is 3 and
// MACRO is 4 in libratbag itself.
const (
	actionNone    uint32 = 0
	actionButton  uint32 = 1
	actionSpecial uint32 = 2
	actionKey     uint32 = 3
	actionMacro   uint32 = 4
	actionUnknown uint32 = 1000
)

// libratbag macro event types (enum ratbag_macro_event_type).
const (
	macroPress   uint32 = 1
	macroRelease uint32 = 2
	macroWait    uint32 = 3
)

// MacroEvent is one step of a recorded macro.
type MacroEvent struct {
	Type    uint32 `json:"type"`
	Keycode uint32 `json:"keycode"`
}

// Resolution is a DPI slot on a profile.
type Resolution struct {
	Path    string   `json:"path"`
	Index   uint32   `json:"index"`
	DPI     uint32   `json:"dpi"`
	DPIs    []uint32 `json:"dpis"`
	Active  bool     `json:"active"`
	Default bool     `json:"default"`
}

// Button is one physical button and its current mapping.
type Button struct {
	Path        string       `json:"path"`
	Index       uint32       `json:"index"`
	ActionType  uint32       `json:"actionType"`
	ActionTypes []uint32     `json:"actionTypes"`
	Value       uint32       `json:"value"`
	Macro       []MacroEvent `json:"macro"`
}

// Profile is one onboard profile.
type Profile struct {
	Path        string       `json:"path"`
	Index       uint32       `json:"index"`
	Name        string       `json:"name"`
	Enabled     bool         `json:"enabled"`
	Active      bool         `json:"active"`
	ReportRate  uint32       `json:"reportRate"`
	ReportRates []uint32     `json:"reportRates"`
	Resolutions []Resolution `json:"resolutions"`
	Buttons     []Button     `json:"buttons"`
}

// Device is a ratbagd device and its full tree.
type Device struct {
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	Model    string    `json:"model"`
	Profiles []Profile `json:"profiles"`
}

// Status describes the daemon connection for the UI.
type Status struct {
	Connected  bool   `json:"connected"`
	APIVersion int32  `json:"apiVersion"`
	Error      string `json:"error"`
}
