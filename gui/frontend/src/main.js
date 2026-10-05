import './style.css';
import './app.css';

import {
    Status,
    ListDevices,
    SetResolution,
    SetResolutionActive,
    SetResolutionDefault,
    SetReportRate,
    SetProfileActive,
    SetProfileEnabled,
    SetProfileName,
    SetButtonAction,
    SetButtonMacro,
    DisableButton,
    SaveBackup,
    ListBackups,
    RestoreBackup,
    DeleteBackup,
} from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime';

import { ACTION, ACTION_LABEL, MACRO, SPECIALS, specialLabel, BUTTON_LABELS } from './spec';
import { CODE_TO_EVDEV, evdevName, describeMacro } from './keymap';

const state = {
    status: null,
    devices: [],
    deviceIndex: 0,
    profileIndex: 0,
    editor: null,
    recording: false,
    busy: false,
    backups: [],
    log: ['Ready.'],
};

const $ = (id) => document.getElementById(id);

function escapeHtml(value) {
    return String(value ?? '')
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;');
}

function log(message) {
    const stamp = new Date().toLocaleTimeString();
    state.log.unshift(`[${stamp}] ${message}`);
    state.log = state.log.slice(0, 200);
    const el = $('log');
    if (el) el.textContent = state.log.join('\n');
}

const currentDevice = () => state.devices[state.deviceIndex] || null;
const currentProfile = () => {
    const device = currentDevice();
    return device ? device.profiles[state.profileIndex] : null;
};

function clampIndices() {
    if (state.deviceIndex >= state.devices.length) state.deviceIndex = 0;
    const device = currentDevice();
    if (!device || device.profiles.length === 0) {
        state.profileIndex = 0;
        state.editor = null;
        return;
    }
    if (state.profileIndex >= device.profiles.length) state.profileIndex = 0;

    if (state.editor) {
        const profile = currentProfile();
        const button = profile?.buttons.find((b) => b.index === state.editor.buttonIndex);
        if (!button) state.editor = null;
    }
}

async function reload() {
    try {
        state.devices = await ListDevices();
        clampIndices();
    } catch (err) {
        log(`Reload failed: ${err}`);
    }
}

async function loadBackups() {
    try {
        state.backups = await ListBackups();
    } catch (err) {
        log(`Load backups failed: ${err}`);
    }
}

async function withBusy(label, fn) {
    if (state.busy) return;
    state.busy = true;
    render();

    try {
        await fn();
        await reload();
        await loadBackups();
        log(`${label}: applied`);
    } catch (err) {
        log(`${label}: ERROR ${err}`);
    } finally {
        state.busy = false;
        render();
    }
}

// --- Rendering ---------------------------------------------------------------

function render() {
    const connected = state.status?.connected;
    document.querySelector('#app').innerHTML = `
        <header>
            <h1>G502 X Config</h1>
            <p class="sub">Configures the mouse through <code>ratbagd</code> over D-Bus.</p>
        </header>
        ${renderStatus()}
        ${connected ? renderDevices() : ''}
        ${connected ? renderBackups() : ''}
        ${renderLog()}
    `;
    attachHandlers();
}

function renderStatus() {
    const status = state.status;
    if (!status) {
        return `<section class="card"><p class="status">Connecting to ratbagd…</p></section>`;
    }
    if (status.connected) {
        return `<section class="card"><p class="status ok">
            Connected to ratbagd · D-Bus API v${status.apiVersion}
        </p></section>`;
    }
    return `<section class="card"><p class="status bad">
        ${escapeHtml(status.error || 'ratbagd is not available')}
    </p></section>`;
}

function renderDevices() {
    const device = currentDevice();
    if (!device) {
        return `<section class="card"><p class="status">No libratbag devices found.</p></section>`;
    }

    const options = state.devices
        .map((d, i) =>
            `<option value="${i}" ${i === state.deviceIndex ? 'selected' : ''}>${escapeHtml(d.name)} — ${escapeHtml(d.model)}</option>`)
        .join('');

    return `
        <section class="card">
            <div class="row">
                <label for="device">Device</label>
                <select id="device">${options}</select>
            </div>
            <p class="status">${device.profiles.length} profiles</p>
        </section>
        <section class="card">
            <div class="tabs">
                ${device.profiles.map((p, i) => `
                    <button class="tab ${i === state.profileIndex ? 'active' : ''}" data-profile="${i}">
                        P${p.index}${p.active ? ' •' : ''}${p.enabled ? '' : ' (off)'}
                    </button>`).join('')}
            </div>
        </section>
        ${renderProfile(currentProfile())}
    `;
}

function renderProfile(profile) {
    if (!profile) return '';

    const nameDisabled = profile.name === '' ? 'disabled' : '';
    const rateOptions = profile.reportRates
        .map((r) => `<option value="${r}" ${r === profile.reportRate ? 'selected' : ''}>${r} Hz</option>`)
        .join('');

    return `
        <section class="card">
            <div class="row spread">
                <h2>Profile ${profile.index}</h2>
                <span class="badge ${profile.active ? 'ok' : ''}">${profile.active ? 'active' : 'inactive'}</span>
            </div>
            <div class="row">
                <label for="profile-name">Name</label>
                <input id="profile-name" type="text" value="${escapeHtml(profile.name)}"
                       placeholder="(device-managed)" ${nameDisabled}/>
                <button id="save-name" class="btn" ${nameDisabled}>Save</button>
            </div>
            <div class="row">
                <label for="profile-enabled">Enabled</label>
                <input id="profile-enabled" type="checkbox" ${profile.enabled ? 'checked' : ''}/>
                <button id="activate" class="btn" ${profile.active ? 'disabled' : ''}>Set active</button>
            </div>
            <div class="row">
                <label for="report-rate">Report rate</label>
                <select id="report-rate">${rateOptions}</select>
            </div>
        </section>
        ${renderResolutions(profile)}
        ${renderButtons(profile)}
    `;
}

function renderResolutions(profile) {
    const rows = profile.resolutions.map((res) => {
        const values = res.dpis && res.dpis.length ? res.dpis : [res.dpi];
        const options = values
            .map((dpi) => `<option value="${dpi}" ${dpi === res.dpi ? 'selected' : ''}>${dpi}</option>`)
            .join('');
        return `
            <tr>
                <td>${res.index}
                    ${res.active ? '<span class="badge ok">active</span>' : ''}
                    ${res.default ? '<span class="badge">default</span>' : ''}
                </td>
                <td><select class="dpi-select" data-res="${escapeHtml(res.path)}">${options}</select></td>
                <td><button class="btn small set-active" data-res="${escapeHtml(res.path)}"
                        ${res.active ? 'disabled' : ''}>Active</button></td>
                <td><button class="btn small set-default" data-res="${escapeHtml(res.path)}"
                        ${res.default ? 'disabled' : ''}>Default</button></td>
            </tr>`;
    }).join('');

    return `
        <section class="card">
            <h2>Resolutions</h2>
            <table class="table">
                <thead><tr><th>Slot</th><th>DPI</th><th></th><th></th></tr></thead>
                <tbody>${rows}</tbody>
            </table>
        </section>
    `;
}

function describeButton(button) {
    switch (button.actionType) {
        case ACTION.NONE: return '<span class="muted">Disabled</span>';
        case ACTION.BUTTON: return `Mouse button ${button.value}`;
        case ACTION.SPECIAL: return escapeHtml(specialLabel(button.value));
        case ACTION.KEY: return `Key: ${escapeHtml(evdevName(button.value))}`;
        case ACTION.MACRO: return `Macro: ${escapeHtml(describeMacro(button.macro))}`;
        default: return `<span class="muted">Unknown (${button.actionType})</span>`;
    }
}

function renderButtons(profile) {
    const rows = profile.buttons.map((button) => `
        <tr class="${state.editor?.buttonIndex === button.index ? 'selected' : ''}">
            <td>${button.index} <span class="muted">${escapeHtml(BUTTON_LABELS[button.index] || '')}</span></td>
            <td>${describeButton(button)}</td>
            <td><button class="btn small edit-btn" data-button="${button.index}">Edit</button></td>
        </tr>`).join('');

    return `
        <section class="card">
            <h2>Buttons</h2>
            <table class="table">
                <tbody>${rows}</tbody>
            </table>
            ${state.editor ? renderEditor(profile) : ''}
        </section>
    `;
}

function renderEditor(profile) {
    const editor = state.editor;
    const button = profile.buttons.find((b) => b.index === editor.buttonIndex);
    const types = (button.actionTypes.length ? button.actionTypes : [0, 1, 2, 3, 4])
        .filter((t) => t < ACTION.UNKNOWN);

    const typeOptions = types
        .map((t) => `<option value="${t}" ${t === editor.actionType ? 'selected' : ''}>${ACTION_LABEL[t] || t}</option>`)
        .join('');

    return `
        <div class="editor">
            <div class="row spread">
                <h3>Button ${editor.buttonIndex} — ${escapeHtml(BUTTON_LABELS[editor.buttonIndex] || '')}</h3>
                <button id="close-editor" class="btn small">Close</button>
            </div>
            <div class="row">
                <label for="action-type">Action</label>
                <select id="action-type">${typeOptions}</select>
            </div>
            ${renderActionEditor(editor)}
            <div class="row spread">
                <span class="muted">${state.recording
                    ? 'Recording… press keys, Esc or Stop to finish.'
                    : 'Changes are written to the mouse immediately.'}</span>
                <button id="apply-action" class="btn">Apply to button</button>
            </div>
        </div>
    `;
}

function renderActionEditor(editor) {
    switch (editor.actionType) {
        case ACTION.NONE:
            return `<p class="muted">The button will send nothing.</p>`;

        case ACTION.BUTTON:
            return `<div class="row">
                <label for="button-number">Mouse button #</label>
                <input id="button-number" type="number" min="0" max="10" value="${editor.value}"/>
            </div>`;

        case ACTION.SPECIAL: {
            const options = SPECIALS
                .map((s) => `<option value="${s.value}" ${s.value === editor.value ? 'selected' : ''}>${escapeHtml(s.label)}</option>`)
                .join('');
            return `<div class="row">
                <label for="special-select">Special action</label>
                <select id="special-select">${options}</select>
            </div>`;
        }

        case ACTION.KEY:
            return `<div class="row">
                <label>Key</label>
                <span class="capture">${editor.value ? escapeHtml(evdevName(editor.value)) : 'none'}</span>
                <button id="record" class="btn small ${state.recording ? 'recording' : ''}">
                    ${state.recording ? 'Stop' : 'Capture key'}
                </button>
            </div>`;

        case ACTION.MACRO:
            return `<div class="row">
                <label>Sequence</label>
                <span class="capture">${escapeHtml(describeMacro(editor.macro))}</span>
                <button id="record" class="btn small ${state.recording ? 'recording' : ''}">
                    ${state.recording ? 'Stop' : 'Record'}
                </button>
                <button id="clear-capture" class="btn small">Clear</button>
            </div>`;

        default:
            return '';
    }
}

function formatDate(value) {
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

function renderBackups() {
    const rows = state.backups.length
        ? state.backups.map((backup) => `
            <tr>
                <td>${escapeHtml(formatDate(backup.createdAt))}<br>
                    <span class="muted">${escapeHtml(backup.file)} · ${backup.profiles} profiles</span>
                </td>
                <td><button class="btn small restore-backup" data-path="${escapeHtml(backup.path)}">Restore</button></td>
                <td><button class="btn small delete-backup" data-path="${escapeHtml(backup.path)}">Delete</button></td>
            </tr>`).join('')
        : `<tr><td class="muted">No backups yet.</td><td></td><td></td></tr>`;

    return `
        <section class="card">
            <div class="row spread">
                <h2>Backup &amp; restore</h2>
                <button id="save-backup" class="btn">Save current settings</button>
            </div>
            <p class="status">Stored in <code>~/.config/ratbegger-g502x/backups</code>.
                Restoring snapshots the current state first, so it can be undone.</p>
            <table class="table">
                <tbody>${rows}</tbody>
            </table>
        </section>
    `;
}

function renderLog() {
    return `<section class="card"><h2>Log</h2><pre id="log" class="output">${escapeHtml(state.log.join('\n'))}</pre></section>`;
}

// --- Key recording -----------------------------------------------------------

function startRecording() {
    state.recording = true;
    if (state.editor.actionType === ACTION.MACRO) {
        state.editor.macro = [];
    }
    window.addEventListener('keydown', onRecordKeyDown, true);
    window.addEventListener('keyup', onRecordKeyUp, true);
    render();
}

function stopRecording() {
    state.recording = false;
    window.removeEventListener('keydown', onRecordKeyDown, true);
    window.removeEventListener('keyup', onRecordKeyUp, true);
    render();
}

function onRecordKeyDown(event) {
    event.preventDefault();
    event.stopPropagation();

    if (event.code === 'Escape') {
        stopRecording();
        return;
    }
    if (event.repeat) return;

    const keycode = CODE_TO_EVDEV[event.code];
    if (keycode === undefined) {
        log(`Key "${event.code}" has no evdev mapping`);
        return;
    }

    if (state.editor.actionType === ACTION.KEY) {
        state.editor.value = keycode;
    } else {
        state.editor.macro.push({ type: MACRO.PRESS, keycode });
    }
    render();
}

function onRecordKeyUp(event) {
    event.preventDefault();
    event.stopPropagation();

    if (state.editor.actionType === ACTION.KEY) return;

    const keycode = CODE_TO_EVDEV[event.code];
    if (keycode === undefined) return;

    state.editor.macro.push({ type: MACRO.RELEASE, keycode });
    render();
}

// --- Actions -----------------------------------------------------------------

function openEditor(index) {
    const profile = currentProfile();
    const button = profile.buttons.find((b) => b.index === index);
    state.editor = {
        buttonIndex: index,
        actionType: button.actionType < ACTION.UNKNOWN ? button.actionType : ACTION.NONE,
        value: button.value,
        macro: [...(button.macro || [])],
    };
    state.recording = false;
    render();
}

async function applyEditor() {
    const device = currentDevice();
    const profile = currentProfile();
    const editor = state.editor;
    if (!device || !profile || !editor) return;

    const button = profile.buttons.find((b) => b.index === editor.buttonIndex);
    if (!button) return;

    switch (editor.actionType) {
        case ACTION.NONE:
            await withBusy('Disable button', () => DisableButton(device.path, button.path));
            break;
        case ACTION.BUTTON:
        case ACTION.SPECIAL:
        case ACTION.KEY:
            await withBusy('Map button', () =>
                SetButtonAction(device.path, button.path, editor.actionType, editor.value >>> 0));
            break;
        case ACTION.MACRO:
            await withBusy('Map macro', () =>
                SetButtonMacro(device.path, button.path, editor.macro));
            break;
        default:
            break;
    }
}

// --- Event wiring ------------------------------------------------------------

function attachHandlers() {
    const deviceSelect = $('device');
    if (deviceSelect) {
        deviceSelect.onchange = () => {
            state.deviceIndex = Number(deviceSelect.value);
            state.profileIndex = 0;
            state.editor = null;
            render();
        };
    }

    document.querySelectorAll('[data-profile]').forEach((el) => {
        el.onclick = () => {
            state.profileIndex = Number(el.dataset.profile);
            state.editor = null;
            render();
        };
    });

    const activate = $('activate');
    if (activate) {
        activate.onclick = () => {
            const device = currentDevice();
            const profile = currentProfile();
            withBusy('Set active profile', () => SetProfileActive(device.path, profile.path));
        };
    }

    const enabled = $('profile-enabled');
    if (enabled) {
        enabled.onchange = () => {
            const device = currentDevice();
            const profile = currentProfile();
            withBusy('Toggle profile', () => SetProfileEnabled(device.path, profile.path, enabled.checked));
        };
    }

    const saveName = $('save-name');
    if (saveName) {
        saveName.onclick = () => {
            const device = currentDevice();
            const profile = currentProfile();
            withBusy('Rename profile', () => SetProfileName(device.path, profile.path, $('profile-name').value));
        };
    }

    const rate = $('report-rate');
    if (rate) {
        rate.onchange = () => {
            const device = currentDevice();
            const profile = currentProfile();
            withBusy('Set report rate', () => SetReportRate(device.path, profile.path, Number(rate.value)));
        };
    }

    document.querySelectorAll('.dpi-select').forEach((el) => {
        el.onchange = () => {
            const device = currentDevice();
            withBusy('Set DPI', () => SetResolution(device.path, el.dataset.res, Number(el.value)));
        };
    });

    document.querySelectorAll('.set-active').forEach((el) => {
        el.onclick = () => {
            const device = currentDevice();
            withBusy('Set active resolution', () => SetResolutionActive(device.path, el.dataset.res));
        };
    });

    document.querySelectorAll('.set-default').forEach((el) => {
        el.onclick = () => {
            const device = currentDevice();
            withBusy('Set default resolution', () => SetResolutionDefault(device.path, el.dataset.res));
        };
    });

    document.querySelectorAll('.edit-btn').forEach((el) => {
        el.onclick = () => openEditor(Number(el.dataset.button));
    });

    const closeEditor = $('close-editor');
    if (closeEditor) {
        closeEditor.onclick = () => {
            if (state.recording) stopRecording();
            state.editor = null;
            render();
        };
    }

    const actionType = $('action-type');
    if (actionType) {
        actionType.onchange = () => {
            const editor = state.editor;
            editor.actionType = Number(actionType.value);
            if (editor.actionType === ACTION.SPECIAL) {
                editor.value = SPECIALS[0].value;
            } else if (editor.actionType === ACTION.MACRO) {
                editor.macro = editor.macro || [];
            } else {
                editor.value = 0;
            }
            render();
        };
    }

    const buttonNumber = $('button-number');
    if (buttonNumber) {
        buttonNumber.onchange = () => { state.editor.value = Number(buttonNumber.value); };
    }

    const specialSelect = $('special-select');
    if (specialSelect) {
        specialSelect.onchange = () => { state.editor.value = Number(specialSelect.value); };
    }

    const record = $('record');
    if (record) {
        record.onclick = () => { state.recording ? stopRecording() : startRecording(); };
    }

    const clear = $('clear-capture');
    if (clear) {
        clear.onclick = () => {
            state.editor.macro = [];
            render();
        };
    }

    const apply = $('apply-action');
    if (apply) {
        apply.onclick = applyEditor;
    }

    const saveBackup = $('save-backup');
    if (saveBackup) {
        saveBackup.onclick = () => {
            const device = currentDevice();
            withBusy('Save backup', () => SaveBackup(device ? device.model : ''));
        };
    }

    document.querySelectorAll('.restore-backup').forEach((el) => {
        el.onclick = () => {
            const file = el.dataset.path.split(/[\\/]/).pop();
            if (!confirm(`Restore ${file}? This overwrites the mouse's current settings.`)) return;
            withBusy('Restore backup', () => RestoreBackup(el.dataset.path));
        };
    });

    document.querySelectorAll('.delete-backup').forEach((el) => {
        el.onclick = () => {
            const file = el.dataset.path.split(/[\\/]/).pop();
            if (!confirm(`Delete ${file}?`)) return;
            withBusy('Delete backup', () => DeleteBackup(el.dataset.path));
        };
    });
}

// --- Boot --------------------------------------------------------------------

EventsOn('ratbagd:resync', (devicePath) => {
    log(`Device resynced by ratbagd: ${devicePath}`);
    reload().then(render);
});

(async () => {
    try {
        state.status = await Status();
    } catch (err) {
        state.status = { connected: false, error: String(err) };
    }

    if (state.status.connected) {
        await reload();
        await loadBackups();
    }
    render();
})();
