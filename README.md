# LOGITECH G502 X

<img src="g502x.png" alt="Logitech G502 X">

## INTRO

The intention of the ratbegger.sh script is to make it easier to load different profiles onto the mouse and set the active one. The script relies upon one or more profile configuration files which are stored in the 'profiles' directory. A sample configuration file is included.

The script uses ratbagctl which is provided by libratbag, a required dependency. The script does not perform all of the functions available with ratbagctl.

## BASIC USAGE

Mouse buttons typically start with the number 1, however libratbag starts with button 0 which is equivalent to mouse button 1, the left primary button. See the image above for reference when writing configuration files.

Start by creating a configuration profile in the 'profiles' directory for the mouse. See the included example file for reference.

When mapping mouse buttons to keyboard keys, a valid key name must be used, all of which start with "KEY_". See the tables below. If you want to map a button to a key which isn't included in the tables, you will need to locate your key code file in order to find the key name. You can try looking in `/usr/include/linux/input-event-codes.h`.

You can have as many configuration profiles as you want in the 'profiles' directory, but only 5 can be loaded on the mouse. The files must have a '.ini' extension. See the included sample configuration file for reference.

To run the script, `cd` to the script directory and run `./ratbegger.sh`.

## HELPFUL HINTS

Following are some examples of how to write macros:

* press and release '1', press 'Left Ctrl', press and release 'A', release 'Left Ctrl': `KEY_1 +KEY_LEFTCTRL KEY_A -KEY_LEFTCTRL`
* press 'A', pause 1 second, release 'A': `+KEY_A t1000 -KEY_A`

I'm not sure how changing the DPI for the 'sniper' button (typically button 4) is supposed to be programed. There's a special action named `resolution-alternate`, but i don't how to use it. One way to change the DPI only while the assigned button is pressed is to have 2 adjacent and identical profiles, the exception being the default DPI setting for the 'sniper' profile, then mapping the 'sniper' button to `profile-cycle-up` or `profile-cycle-down`. The following example assumes profile 0 is the default and profile 1 will be used to lower DPI while the 'sniper' button is pressed:

Default (profile 0) resolution setting:

```
...
default resolution profile=0
active resolution profile=0
resolution 0=1600
...
button 4=profile-cycle-up
...
```

Sniper (profile 1) resolution setting:

```
...
default resolution profile=0
active resolution profile=0
resolution 0=800
...
```

## KNOWN ISSUES

* Although `profile name=Some Name` can be used in configuration files to assign a name to a profile, there's a bug in libratbag that prevents writing the profile name to the mouse, therefore these lines are commented out in the sample configuration files (see: https://github.com/libratbag/libratbag/issues/680).
* Not all mouse buttons can be mapped to a macro. This seems to be either a libratbag or mouse firmware bug.
* There are several other issues with libratbag and the G502 series of mice.

## G502 X AVAILABLE RESOLUTION/DPI SETTINGS

100 to 1000 in increments of 50, 1000 to 2600 in increments of 100, 2600 to 5000 in increments of 200, 5000 to 25500 in increments of 500

## G502 X AVAILABLE USB POLLING RATES (Hz)

125 250 500 1000

## G502 X SPECIAL ACTIONS

The following are special actions which can be mapped to mouse buttons. If you use any of the up/down actions in a configuration file, you may want to duplicate them in any other configuration files.

|      DESCRIPTION      |  SPECIAL ACTION NAME  |
| --------------------- | --------------------- |
| disable button        | unknown
| increase profile      | profile-up            |
| decrease profile      | profile-down          |
| cycle profile up      | profile-cycle-up      |
| cycle profile down    | profile-cycle-down    |
| increase resolution   | resolution-up         |
| decrease resolution   | resolution-down       |
| cycle resolution up   | resolution-cycle-up   |
| cycle resolution down | resolution-cycle-down |
| default resolution    | resolution-default    |
| sniper resolution     | resolution-alternate  |
| double click          | doubleclick           |
| wheel tilt left       | wheel-left            |
| wheel tilt right      | wheel-right           |
| wheel up              | wheel-up              |
| wheel down            | wheel-down            |

## COMMON KEYBOARD KEYS

|     DESCRIPTION      |    KEY NAME    |
| -------------------- | -------------- |
| Function Keys F1-F12 | KEY_F[1-12]    |
| Number Keys 0-9      | KEY_[0-9]      |
| Letter Keys A-Z      | KEY_[A-Z]      |
| Apostrophe           | KEY_APOSTROPHE |
| Backslash            | KEY_BACKSLASH  |
| Backspace            | KEY_BACKSPACE  |
| Caps Lock            | KEY_CAPSLOCK   |
| Comma                | KEY_COMMA      |
| Delete               | KEY_DELETE     |
| Dot                  | KEY_DOT        |
| Down                 | KEY_DOWN       |
| End                  | KEY_END        |
| Enter                | KEY_ENTER      |
| Equals               | KEY_EQUAL      |
| Escape               | KEY_ESC        |
| Forward Slash        | KEY_SLASH      |
| Grave                | KEY_GRAVE      |
| Home                 | KEY_HOME       |
| Insert               | KEY_INSERT     |
| Left                 | KEY_LEFT       |
| Left Alt             | KEY_LEFTALT    |
| Left Brace           | KEY_LEFTBRACE  |
| Left Control         | KEY_LEFTCTRL   |
| Left Meta            | KEY_LEFTMETA   |
| Left Shift           | KEY_LEFTSHIFT  |
| Minus                | KEY_MINUS      |
| Num Lock             | KEY_NUMLOCK    |
| Page Down            | KEY_PAGEDOWN   |
| Page Up              | KEY_PAGEUP     |
| Pause                | KEY_PAUSE      |
| Right                | KEY_RIGHT      |
| Right Alt            | KEY_RIGHTALT   |
| Right Brace          | KEY_RIGHTBRACE |
| Right Control        | KEY_RIGHTCTRL  |
| Right Meta           | KEY_RIGHTMETA  |
| Right Shift          | KEY_RIGHTSHIFT |
| Scroll Lock          | KEY_SCROLLLOCK |
| Semicolon            | KEY_SEMICOLON  |
| Space                | KEY_SPACE      |
| Tab                  | KEY_TAB        |
| Up                   | KEY_UP         |

## KEYPAD KEYS

|   DESCRIPTION   |    KEY NAME    |
| --------------- | -------------- |
| Asterisk        | KEY_KPASTERISK |
| Comma           | KEY_KPCOMMA    |
| Dot             | KEY_KPDOT      |
| Enter           | KEY_KPENTER    |
| Equals          | KEY_KPEQUAL    |
| Forward Slash   | KEY_KPSLASH    |
| Minus           | KEY_KPMINUS    |
| Number Keys 0-9 | KEY_KP0-9      |
| Plus            | KEY_KPPLUS     |

## MEDIA KEYS

| DESCRIPTION |     KEY NAME     |
| ----------- | ---------------- |
| Back        | KEY_BACK         |
| Forward     | KEY_FORWARD      |
| Mute        | KEY_MUTE         |
| Next        | KEY_NEXTSONG     |
| Play/Pause  | KEY_PLAYPAUSE    |
| Previous    | KEY_PREVIOUSSONG |
| Stop        | KEY_STOP         |
| Volume Down | KEY_VOLUMEDOWN   |
| Volume Up   | KEY_VOLUMEUP     |

## RATBAGCTL COMMAND EXAMPLES

Following are some examples of ratbagctl commands if you need to use it directly.

* list devices: `ratbagctl list`
* get device info: `ratbagctl 'Logitech G502 X' info`
* set USB polling rate: `ratbagctl 'Logitech G502 X' profile <0-4> rate set <125|250|500|1000>`
* enable a profile: `ratbagctl 'Logitech G502 X' profile <0-4> enable`
* disable a profile: `ratbagctl 'Logitech G502 X' profile <0-4> disable`
* set currently active profile: `ratbagctl 'Logitech G502 X' profile active set <0-4>`
* set default resolution profile: `ratbagctl 'Logitech G502 X' profile <0-4> resolution default set <0-4>`
* set active resolution profile: `ratbagctl 'Logitech G502 X' profile <0-4> resolution active set <0-4>`
* set DPI for given resolution profile: `ratbagctl 'Logitech G502 X' profile <0-4> resolution <0-4> dpi set <see table above>`
* map one mouse button to another: `ratbagctl 'Logitech G502 X' profile <0-4> button <0-10> action set button <0-10>`
* map mouse button to keyboard key: `ratbagctl 'Logitech G502 X' profile <0-4> button <0-10> action set macro <see table above>`
* map mouse button to a special action: `ratbagctl 'Logitech G502 X' profile <0-4> button <0-10> action set special <see table above>`

## RESOURCES

* RatBegger-G502X source code repository: https://codeberg.org/12bytes/ratbegger-g502x
* libratbag source code repository: https://github.com/libratbag/libratbag
