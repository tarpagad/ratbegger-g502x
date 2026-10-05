// Browser KeyboardEvent.code -> Linux evdev keycode (linux/input-event-codes.h).
// These are the values libratbag stores in button mappings and macros.
export const CODE_TO_EVDEV = {
  // Letters
  KeyA: 30, KeyB: 48, KeyC: 46, KeyD: 32, KeyE: 18, KeyF: 33, KeyG: 34,
  KeyH: 35, KeyI: 23, KeyJ: 36, KeyK: 37, KeyL: 38, KeyM: 50, KeyN: 49,
  KeyO: 24, KeyP: 25, KeyQ: 16, KeyR: 19, KeyS: 31, KeyT: 20, KeyU: 22,
  KeyV: 47, KeyW: 17, KeyX: 45, KeyY: 21, KeyZ: 44,
  // Digits
  Digit1: 2, Digit2: 3, Digit3: 4, Digit4: 5, Digit5: 6, Digit6: 7,
  Digit7: 8, Digit8: 9, Digit9: 10, Digit0: 11,
  // Control and punctuation
  Escape: 1, Minus: 12, Equal: 13, Backspace: 14, Tab: 15, BracketLeft: 26,
  BracketRight: 27, Enter: 28, ControlLeft: 29, Semicolon: 39, Quote: 40,
  Backquote: 41, ShiftLeft: 42, Backslash: 43, Comma: 51, Period: 52,
  Slash: 53, ShiftRight: 54, AltLeft: 56, Space: 57, CapsLock: 58,
  ControlRight: 97, AltRight: 100, MetaLeft: 125, MetaRight: 126,
  ContextMenu: 127,
  // Function keys
  F1: 59, F2: 60, F3: 61, F4: 62, F5: 63, F6: 64, F7: 65, F8: 66,
  F9: 67, F10: 68, F11: 87, F12: 88,
  // Navigation
  Home: 102, ArrowUp: 103, PageUp: 104, ArrowLeft: 105, ArrowRight: 106,
  End: 107, ArrowDown: 108, PageDown: 109, Insert: 110, Delete: 111,
  PrintScreen: 99, Pause: 119, ScrollLock: 70,
  // Numpad
  NumLock: 69, NumpadDivide: 98, NumpadMultiply: 55, NumpadSubtract: 74,
  NumpadAdd: 78, NumpadEnter: 96, NumpadEqual: 117, NumpadDecimal: 83,
  Numpad0: 82, Numpad1: 79, Numpad2: 80, Numpad3: 81, Numpad4: 75,
  Numpad5: 76, Numpad6: 77, Numpad7: 71, Numpad8: 72, Numpad9: 73,
  // Media
  AudioVolumeMute: 113, AudioVolumeDown: 114, AudioVolumeUp: 115,
  MediaTrackNext: 163, MediaPlayPause: 164, MediaTrackPrevious: 165,
  MediaStop: 166, BrowserBack: 158, BrowserForward: 159, BrowserHome: 172,
};

// Human labels keyed by evdev keycode.
const KEY_LABEL = {
  1: 'Esc', 2: '1', 3: '2', 4: '3', 5: '4', 6: '5', 7: '6', 8: '7', 9: '8',
  10: '9', 11: '0', 12: '-', 13: '=', 14: 'Backspace', 15: 'Tab', 16: 'Q',
  17: 'W', 18: 'E', 19: 'R', 20: 'T', 21: 'Y', 22: 'U', 23: 'I', 24: 'O',
  25: 'P', 26: '[', 27: ']', 28: 'Enter', 29: 'Ctrl', 30: 'A', 31: 'S',
  32: 'D', 33: 'F', 34: 'G', 35: 'H', 36: 'J', 37: 'K', 38: 'L', 39: ';',
  40: "'", 41: '`', 42: 'Shift', 43: '\\', 44: 'Z', 45: 'X', 46: 'C',
  47: 'V', 48: 'B', 49: 'N', 50: 'M', 51: ',', 52: '.', 53: '/', 54: 'Shift',
  55: 'Num*', 56: 'Alt', 57: 'Space', 58: 'CapsLock', 59: 'F1', 60: 'F2',
  61: 'F3', 62: 'F4', 63: 'F5', 64: 'F6', 65: 'F7', 66: 'F8', 67: 'F9',
  68: 'F10', 69: 'NumLock', 70: 'ScrollLock', 71: 'Num7', 72: 'Num8',
  73: 'Num9', 74: 'Num-', 75: 'Num4', 76: 'Num5', 77: 'Num6', 78: 'Num+',
  79: 'Num1', 80: 'Num2', 81: 'Num3', 82: 'Num0', 83: 'Num.', 87: 'F11',
  88: 'F12', 96: 'NumEnter', 97: 'RCtrl', 98: 'Num/', 99: 'PrtSc', 100: 'RAlt',
  102: 'Home', 103: 'Up', 104: 'PgUp', 105: 'Left', 106: 'Right', 107: 'End',
  108: 'Down', 109: 'PgDn', 110: 'Insert', 111: 'Delete', 113: 'Mute',
  114: 'Vol-', 115: 'Vol+', 117: 'Num=', 119: 'Pause', 125: 'Meta',
  126: 'RMeta', 127: 'Menu', 158: 'Back', 159: 'Forward', 163: 'Next',
  164: 'Play', 165: 'Prev', 166: 'Stop', 172: 'Home',
};

const MODIFIER_KEYS = new Set([29, 97, 42, 54, 56, 100, 125, 126]);

export function evdevName(keycode) {
  return KEY_LABEL[keycode] || `KEY_${keycode}`;
}

// Render a recorded macro as a readable string, e.g. "Ctrl+C" or "A B".
export function describeMacro(macro) {
  if (!macro || macro.length === 0) {
    return 'empty';
  }

  const chords = [];
  let modifiers = [];
  let plain = [];

  const flush = () => {
    const keys = [...modifiers, ...plain];
    if (keys.length) {
      chords.push(keys.map(evdevName).join('+'));
    }
    modifiers = [];
    plain = [];
  };

  for (const event of macro) {
    if (event.type === 1) {
      if (MODIFIER_KEYS.has(event.keycode)) {
        modifiers.push(event.keycode);
      } else {
        plain.push(event.keycode);
      }
    } else if (event.type === 2 && !MODIFIER_KEYS.has(event.keycode)) {
      flush();
    }
  }
  flush();

  return chords.join(' ') || 'empty';
}
