# LOGITECH G502 X — ADVANCED USAGE

<img src="g502x.png" alt="Logitech G502 X button numbering">

## RATBAGCTL GUIDE

`ratbagctl` is bundled with [libratbag](https://github.com/libratbag/libratbag) and is the utility that `ratbegger.sh` drives under the hood. This guide configures a brand-new G502 X using `ratbagctl` **directly, with no helper script** — it is the manual equivalent of what the script automates. Follow it once by hand to understand every step, or use it for one-off changes you don't want to add to a configuration file.

Every change is written to the mouse's **onboard memory**, so once you are done, **nothing needs to keep running** for the buttons and DPI to work. The `ratbagd` daemon is only required *while* you are making changes.

> The examples below use a shell variable `DEV` purely for brevity. There is no script involved; every line is an ordinary `ratbagctl` command you can type, adapt, or copy individually.

---

## 1. INSTALL LIBRATBAG

`libratbag` provides both the `ratbagd` daemon and the `ratbagctl` command-line utility. Install it with your distribution's package manager. Package names vary, so verify with your distribution if the first attempt fails:

| Distribution(s)          | Package               | Install command                          |
| ------------------------ | --------------------- | ---------------------------------------- |
| Arch, Manjaro            | `libratbag`           | `sudo pacman -S libratbag`               |
| Debian, Ubuntu, Mint     | `ratbagd`             | `sudo apt install ratbagd`               |
| Fedora                   | `libratbag-ratbagd`   | `sudo dnf install libratbag-ratbagd`     |
| openSUSE                 | `libratbag`           | `sudo zypper install libratbag`          |
| Gentoo                   | `libratbag`           | `sudo emerge libratbag`                  |

Ensure the daemon is running and enabled at boot:

```
sudo systemctl enable --now ratbagd
```

`ratbagd` runs as root and ships a Polkit/udev rule that lets ordinary users talk to it, so you should **not** need `sudo` to run `ratbagctl`. If you get a permission error, see [Troubleshooting](#10-troubleshooting).

---

## 2. IDENTIFY THE DEVICE

List the devices libratbag can see:

```
ratbagctl list
```

This prints a single line per device — first a codename assigned by ratbagd, then the human-readable name:

```
warbling-mara:       Logitech G502 X
```

The codename (here `warbling-mara`) is **generated randomly and can change between sessions**, so it is safer to address the mouse by its real name in quotes, since it contains spaces:

```
DEV='Logitech G502 X'
ratbagctl "$DEV" info
```

Both forms work everywhere in this guide — substitute `warbling-mara` for `'Logitech G502 X'` if you prefer.

---

## 3. UNDERSTAND THE BUTTON NUMBERING

This is the single most confusing part of configuring a Logitech mouse. There are two things to keep in mind:

* Mouse buttons are conventionally numbered starting at **1**, but libratbag numbers them starting at **0**. So libratbag "button 0" is the left (primary) button, "button 1" is the right (secondary) button, and so on.
* The G502 X has **11 buttons**, numbered **0 to 10**, as printed on the included `g502x.png` image.

| Index | Control (see image above)        |
| ----- | -------------------------------- |
| 0     | Left / primary click             |
| 1     | Right / secondary click          |
| 2     | Scroll-wheel click (middle)      |
| 3     | Left-side (thumb) button         |
| 4     | Sniper / DPI-shift button        |
| 5     | Left-side (thumb) button         |
| 6     | Scroll-wheel tilt left           |
| 7     | Scroll-wheel tilt right          |
| 8     | Top button behind the scroll wheel |
| 9     | Left-side (thumb) button         |
| 10    | Left-side (thumb) button         |

To identify a button with certainty, press it and watch the `info` output, or read a single mapping back:

```
ratbagctl "$DEV" button count                 # -> 11
ratbagctl "$DEV" button 9 get                  # show what button 9 is currently mapped to
```

---

## 4. INSPECT THE CURRENT CONFIGURATION

```
ratbagctl "$DEV" info
```

Abridged output from a G502 X:

```
 Logitech G502 X
             Model: usb:046d:c099:0
 Number of Buttons: 11
    Number of Leds: 0
Number of Profiles: 5
Profile 0:
  Report Rate: 1000Hz
  Resolutions:
    0: 800dpi
    1: 1200dpi
    2: 1600dpi (active) (default)
    3: 2400dpi
    4: 3200dpi
  Button: 0 is mapped to 'button 1'
  Button: 1 is mapped to 'button 2'
  ...
Profile 2: (disabled) (active)
Profile 3: (disabled)
Profile 4: (disabled)
```

Things to note:

* **Profiles** — the mouse stores **5 profiles** (0–4) in onboard memory, but only one is *active* at a time. You can only enable/disable and program one profile at a time.
* **Resolutions** — each profile holds **5 DPI slots** (0–4). One slot is the *default* and one is the *active* (current) resolution. The sniper button (`resolution-alternate`) always switches to **resolution 0** while held, so set slot 0 to your slow/sniper DPI.
* **Report Rate** — the USB polling rate.

### A known quirk: profile 0 vs. profile 1

There is a long-standing libratbag bug ([#680](https://github.com/libratbag/libratbag/issues/680)) that makes the profile state confusing:

* Writing to **profile 0** may actually land in **profile 1**, and setting profile 0 active may appear to do nothing.
* A profile can be listed as both `(disabled)` **and** `(active)` at the same time, when it is really a different profile that is active.

The rule of thumb: **work with profile 1** (or explicitly enable the profile you target and re-check), and **always re-run `info` after writing** to confirm the settings landed where you expected. See [Troubleshooting](#10-troubleshooting).

---

## 5. CHOOSE AND ACTIVATE A PROFILE

Enable the profile you want to program (a disabled profile is ignored by the mouse) and make it the active one:

```
ratbagctl "$DEV" profile 1 enable
ratbagctl "$DEV" profile active set 1
```

Confirm with:

```
ratbagctl "$DEV" profile active get
```

Because of the quirk above, if you intend to use profile 0, start from profile 1 or verify afterwards with `info`. It is good practice to program **profile 1** and leave the others disabled.

---

## 6. SET THE RESOLUTIONS (DPI)

DPI commands are scoped to a profile and a resolution slot. The full form is:

```
ratbagctl "$DEV" profile <0-4> resolution <0-4> dpi set <dpi>
```

Example — five DPI steps in profile 1:

```
ratbagctl "$DEV" profile 1 resolution 0 dpi set 800     # sniper slot
ratbagctl "$DEV" profile 1 resolution 1 dpi set 1200
ratbagctl "$DEV" profile 1 resolution 2 dpi set 1600
ratbagctl "$DEV" profile 1 resolution 3 dpi set 2400
ratbagctl "$DEV" profile 1 resolution 4 dpi set 3200
```

Then tell the mouse which slot is the *default* and which is *active*:

```
ratbagctl "$DEV" profile 1 resolution default set 2
ratbagctl "$DEV" profile 1 resolution active set 2
```

### Valid DPI values for the G502 X

Only certain increments are accepted:

* 100 to 1000 DPI — increments of 50
* 1000 to 2600 DPI — increments of 100
* 2600 to 5000 DPI — increments of 200
* 5000 to 25500 DPI — increments of 500

List what your mouse currently reports:

```
ratbagctl "$DEV" dpi get-all
ratbagctl "$DEV" dpi get
```

---

## 7. SET THE USB REPORT RATE

The G502 X supports **125, 250, 500 and 1000 Hz**:

```
ratbagctl "$DEV" profile 1 rate set 1000
```

Check the current rate and the available options:

```
ratbagctl "$DEV" rate get
ratbagctl "$DEV" rate get-all
```

---

## 8. MAP THE BUTTONS

Button commands are scoped to a profile: `profile <N> button <M> action set ...`. There are four kinds of action you can assign.

### a. Map one mouse button to another

```
ratbagctl "$DEV" profile 1 button M action set button B
```

Example — keep the default left/right/middle behaviour:

```
ratbagctl "$DEV" profile 1 button 0 action set button 1
ratbagctl "$DEV" profile 1 button 1 action set button 2
ratbagctl "$DEV" profile 1 button 2 action set button 3
```

### b. Map a button to a special action

```
ratbagctl "$DEV" profile 1 button M action set special <action>
```

Example — use the sniper button and wheel tilt:

```
ratbagctl "$DEV" profile 1 button 4 action set special resolution-alternate
ratbagctl "$DEV" profile 1 button 6 action set special wheel-left
ratbagctl "$DEV" profile 1 button 7 action set special wheel-right
```

The available special actions are listed in the [G502 X SPECIAL ACTIONS](README.md#g502-x-special-actions) table in the README.

### c. Map a button to a keyboard key or macro

```
ratbagctl "$DEV" profile 1 button M action set macro <keys...>
```

Macro syntax (from `man ratbagctl`):

| Token      | Meaning                                |
| ---------- | -------------------------------------- |
| `KEY_A`    | press **and release** the key          |
| `+KEY_A`   | press and **hold** the key             |
| `-KEY_A`   | **release** a held key                 |
| `t300`     | wait 300 ms                            |

Examples:

```
# Button 10 -> Backspace
ratbagctl "$DEV" profile 1 button 10 action set macro KEY_BACKSPACE

# Button 9 -> Ctrl+W (e.g. close a tab)
ratbagctl "$DEV" profile 1 button 9 action set macro +KEY_LEFTCTRL KEY_W -KEY_LEFTCTRL

# Button 8 -> hold A for one second, then release
ratbagctl "$DEV" profile 1 button 8 action set macro +KEY_A t1000 -KEY_A
```

On `ratbagctl` 0.18 and newer you can also bind a single key directly, which is a shortcut for a simple press/release:

```
ratbagctl "$DEV" profile 1 button 10 action set key KEY_BACKSPACE
```

Key names must match the names in your system's input-event-codes header — normally `/usr/include/linux/input-event-codes.h`. Common names are tabulated in the README under [COMMON KEYBOARD KEYS](README.md#common-keyboard-keys), [KEYPAD KEYS](README.md#keypad-keys) and [MEDIA KEYS](README.md#media-keys).

### d. Disable a button

```
ratbagctl "$DEV" profile 1 button M action set disabled
```

(This is the `unknown` special action from the README, in a clearer spelling.)

---

## 9. WRITE, THEN VERIFY

By default **every `ratbagctl` command writes to the mouse immediately**. If you want to stage several changes and commit them together, use `--nocommit` on all but the last command:

```
ratbagctl --nocommit "$DEV" profile 1 rate set 500
ratbagctl --nocommit "$DEV" profile 1 button 9 action set macro +KEY_LEFTCTRL KEY_W -KEY_LEFTCTRL
ratbagctl "$DEV" profile active set 1        # final call writes everything
```

Then verify — this is the step that catches the profile 0/1 quirk:

```
ratbagctl "$DEV" info
```

Confirm the active profile, the report rate, the active/default resolutions, and each `Button: N is mapped to ...` line.

---

## 10. COMPLETE WORKED EXAMPLE

This reproduces the [`configs/sample config.ini`](configs/sample%20config.ini) configuration using only `ratbagctl`, writing to profile 1:

```
DEV='Logitech G502 X'

ratbagctl "$DEV" profile 1 enable
ratbagctl "$DEV" profile 1 rate set 1000

ratbagctl "$DEV" profile 1 resolution 0 dpi set 500
ratbagctl "$DEV" profile 1 resolution 1 dpi set 800
ratbagctl "$DEV" profile 1 resolution 2 dpi set 1100
ratbagctl "$DEV" profile 1 resolution 3 dpi set 1400
ratbagctl "$DEV" profile 1 resolution 4 dpi set 1700
ratbagctl "$DEV" profile 1 resolution default set 2
ratbagctl "$DEV" profile 1 resolution active set 2

ratbagctl "$DEV" profile 1 button 0 action set button 1
ratbagctl "$DEV" profile 1 button 1 action set button 2
ratbagctl "$DEV" profile 1 button 2 action set button 3
ratbagctl "$DEV" profile 1 button 3 action set special resolution-down
ratbagctl "$DEV" profile 1 button 4 action set special resolution-alternate
ratbagctl "$DEV" profile 1 button 5 action set special resolution-up
ratbagctl "$DEV" profile 1 button 6 action set special wheel-left
ratbagctl "$DEV" profile 1 button 7 action set special wheel-right
ratbagctl "$DEV" profile 1 button 8 action set macro KEY_HOME
ratbagctl "$DEV" profile 1 button 9 action set special profile-cycle-up
ratbagctl "$DEV" profile 1 button 10 action set special profile-cycle-down

ratbagctl "$DEV" profile active set 1
ratbagctl "$DEV" info
```

If the `info` output doesn't reflect your changes, re-read [section 5](#5-choose-and-activate-a-profile) — you are almost certainly writing to a different profile than the one that is active.

---

## 11. TROUBLESHOOTING

* **`ratbagctl list` shows nothing / the mouse is missing** — make sure `ratbagd` is running (`systemctl status ratbagd`), then unplug and replug the mouse (or reconnect the USB cable). Check `dmesg` for the device being detected.
* **Permission denied when talking to the daemon** — `ratbagd` must be running, and its udev/Polkit rules installed. Reinstalling the package and replugging the mouse usually fixes it. Avoid running `ratbagctl` under `sudo`, which can create root-owned state.
* **A button doesn't do what you set** — you are likely editing a profile that is not active, or hitting the profile 0/1 quirk. Enable the profile, set it active, and re-check with `info`.
* **A button won't accept a macro** — a known libratbag/G502 limitation; not every button can be mapped to a macro. Try a "special" action or map it to another button instead.
* **Profile name doesn't stick** — a libratbag bug ([#680](https://github.com/libratbag/libratbag/issues/680)); the name is not written to the mouse, which is why `ratbegger.sh` comments out `profile name=` lines.
* **Settings lost after a reboot** — mappings live in onboard memory and should persist; verify with `info` after rebooting. If a profile was never enabled, the mouse may switch to an empty/default one on power-up.

---

## 12. RESOURCES

* `man ratbagctl` (or the [ratbagctl manual page](https://man.archlinux.org/man/extra/libratbag/ratbagctl.1.en))
* [README](README.md) — key-name, special-action and DPI/rate tables, plus the `sample config.ini`
* [ratbegger.sh](ratbegger.sh) — the script that automates everything above
* [libratbag source](https://github.com/libratbag/libratbag)
* [Product review: Logitech G502 X on Linux](https://12bytes.org/product-review-logitech-g502-x-on-linux/) — the article this project grew out of
