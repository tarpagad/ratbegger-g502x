# G502 X Config (Wails GUI)

A desktop configurator for the Logitech G502 X. It talks to **`ratbagd`** — the
[libratbag](https://github.com/libratbag/libratbag) daemon — directly over the system D-Bus, so
the mouse can be configured without typing `ratbagctl` commands (and without spawning a
subprocess). It edits profiles, resolutions and button mappings, including recording keyboard
macros.

> **Status: builds and runs.** Verified on Ubuntu 26.04 with Go 1.26, Wails v2.16 and
> `libwebkit2gtk-4.1-dev`: `wails build -tags webkit2_41` produces `build/bin/gui`, which
> launches and reads the connected G502 X. On distributions that still ship WebKitGTK 4.0, drop
> the tag.

## Prerequisites

- Go 1.25 or newer
- The Wails CLI v2:

  ```sh
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

- Node.js and npm (for the Vite frontend build)
- `libratbag` with a running `ratbagd` (the app is useless without it)
- Linux GUI development libraries:

  | Distribution         | Command                                                            |
  | -------------------- | ----------------------------------------------------------------- |
  | Ubuntu / Debian      | `sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev`             |
  | Fedora               | `sudo dnf install gtk3-devel webkit2gtk4.0-devel`                 |
  | Arch                 | `sudo pacman -S gtk3 webkit2gtk`                                  |

  Newer distributions no longer ship WebKitGTK 4.0. Install `libwebkit2gtk-4.1-dev` (Debian),
  `webkit2gtk4.1-devel` (Fedora) or `webkit2gtk-4.1` (Arch) instead, and pass the
  `webkit2_41` build tag.

## Develop and build

```sh
cd gui
wails dev   -tags webkit2_41     # hot-reload development
wails build -tags webkit2_41     # production binary in build/bin/
```

`wails dev` runs a Vite dev server and exposes a browser dev server on
<http://localhost:34115> where the bound Go methods are also callable.

**If the frontend build fails with `sh: 1: vite: not found`**, `NODE_ENV=production` is set in
your shell, which makes npm skip `devDependencies`. Either build with `NODE_ENV=development`, or
install the toolchain explicitly:

```sh
NODE_ENV=development npm install --include=dev
```

## How it works

- **`ratbagd.go`** — a thin client over `org.freedesktop.ratbag1` using
  [`github.com/godbus/dbus/v5`](https://github.com/godbus/dbus/v5). It reads the device tree
  (Device → Profile → Resolution/Button) and writes individual properties. Object paths are
  always the ones the daemon returns — never constructed locally.
- **`app.go`** — the Wails-bound API. Each setter applies a change and then calls
  `Device.Commit()`, so the frontend never has to remember the D-Bus write protocol.
  `startup` also subscribes to the `Resync` signal and forwards it to the UI.
- **`models.go`** — plain DTOs (`Device`, `Profile`, `Resolution`, `Button`, `MacroEvent`) shared
  with the frontend as JSON.
- **`frontend/src/keymap.js`** — maps browser `KeyboardEvent.code` to Linux evdev keycodes and
  renders macros back into readable strings.
- **`frontend/src/spec.js`** — the button action-type and special-function tables.
- **`frontend/src/main.js`** — the UI: device picker, profile tabs, resolution table, button list
  and the button editor with a key/macro recorder.

## Features

- **Profiles** — switch the active profile, enable/disable, rename (where the device supports it)
  and set the report rate.
- **Resolutions** — edit each DPI slot from the device's permitted value list, and set the active
  and default slots.
- **Buttons** — assign Disabled, a mouse button, a special action, a single key, or a recorded
  macro; the recorder captures press/release events so chords (`Ctrl+C`) and sequences work.

## Notes and gotchas

- **API version must match.** The D-Bus interface has no compatibility guarantees; the client
  checks `Manager.APIVersion` equals `2` and refuses to run otherwise.
- **`Commit` is asynchronous.** ratbagd may emit `Resync` seconds later if a write failed; the UI
  listens for it and reloads the tree.
- **libratbag's published D-Bus docs are wrong about two enums.** The real values are
  `BUTTON_ACTION_TYPE_KEY = 3` and `..._MACRO = 4` (not 3 = Macro), and macro event types are
  `1 = press`, `2 = release`. The code follows the C enums in `src/libratbag-enums.h`, verified
  against a live daemon.
- **Profile enable/disable** is exposed as `Disabled` on libratbag 0.18 and `Enabled` on newer
  versions; the client detects which one is present.
- **Write test on real hardware.** Settings go to the mouse's onboard memory. The integration
  test only reads.

## Testing

```sh
go test ./...                              # unit tests (no daemon needed)
RATBAGD_LIVE=1 go test -run TestLiveDevices -v .   # reads the real device
```

## Known limitations and next steps

- Only reads/writes through `ratbagd`; there is no offline mode.
- LEDs are not exposed (the G502 X has none).
- A clickable mouse diagram and `.ini` import/export (interoperable with `ratbegger.sh`) are the
  obvious next steps.
