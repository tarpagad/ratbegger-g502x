// Button action types. NB: the published libratbag D-Bus docs mislabel these —
// in libratbag KEY is 3 and MACRO is 4.
export const ACTION = {
  NONE: 0,
  BUTTON: 1,
  SPECIAL: 2,
  KEY: 3,
  MACRO: 4,
  UNKNOWN: 1000,
};

export const ACTION_LABEL = {
  [ACTION.NONE]: 'Disabled',
  [ACTION.BUTTON]: 'Mouse button',
  [ACTION.SPECIAL]: 'Special action',
  [ACTION.KEY]: 'Single key',
  [ACTION.MACRO]: 'Macro',
};

// libratbag macro event types.
export const MACRO = {
  PRESS: 1,
  RELEASE: 2,
  WAIT: 3,
};

const special = (offset) => 0x40000000 + offset;

export const SPECIALS = [
  { value: special(0), label: 'Unknown' },
  { value: special(1), label: 'Double click' },
  { value: special(2), label: 'Wheel left' },
  { value: special(3), label: 'Wheel right' },
  { value: special(4), label: 'Wheel up' },
  { value: special(5), label: 'Wheel down' },
  { value: special(6), label: 'Ratchet mode switch' },
  { value: special(7), label: 'Resolution cycle up' },
  { value: special(8), label: 'Resolution cycle down' },
  { value: special(9), label: 'Resolution up' },
  { value: special(10), label: 'Resolution down' },
  { value: special(11), label: 'Resolution alternate (sniper)' },
  { value: special(12), label: 'Resolution default' },
  { value: special(13), label: 'Profile cycle up' },
  { value: special(14), label: 'Profile cycle down' },
  { value: special(15), label: 'Profile up' },
  { value: special(16), label: 'Profile down' },
  { value: special(17), label: 'Second mode' },
  { value: special(18), label: 'Battery level' },
];

export function specialLabel(value) {
  const found = SPECIALS.find((entry) => entry.value === value);
  return found ? found.label : `Special ${value}`;
}

// Physical labels for the G502 X button indices, matching the ADVANCED.md map.
export const BUTTON_LABELS = [
  'Left click',
  'Right click',
  'Middle click',
  'Thumb 1 (side)',
  'Sniper / DPI shift',
  'Thumb 2 (side)',
  'Wheel tilt left',
  'Wheel tilt right',
  'Top button',
  'Thumb 3 (side)',
  'Thumb 4 (side)',
];
