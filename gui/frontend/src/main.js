import './style.css';
import './app.css';

import {
    RatbagctlAvailable,
    ListDevices,
    GetDeviceInfo,
    SetDPI,
    SetReportRate,
    SetActiveProfile,
} from '../wailsjs/go/main/App';

const app = document.querySelector('#app');

app.innerHTML = `
  <header>
    <h1>G502 X Config</h1>
    <p class="sub">A minimal <a href="https://wails.io">Wails</a> front-end for
      <code>ratbagctl</code>. Settings are written to the mouse's onboard memory.</p>
  </header>

  <section class="card">
    <div class="row">
      <label for="device">Device</label>
      <select id="device"></select>
      <button id="refresh" class="btn">Refresh</button>
    </div>
    <p id="status" class="status"></p>
  </section>

  <section class="card">
    <h2>Profile</h2>
    <div class="row">
      <label for="profile">Profile (0-4)</label>
      <input id="profile" type="number" min="0" max="4" value="1"/>
      <button id="activate" class="btn">Set active</button>
    </div>
  </section>

  <section class="card">
    <h2>Resolution &amp; rate</h2>
    <div class="row">
      <label>DPI</label>
      <span class="inline">slot
        <input id="slot" type="number" min="0" max="4" value="2"/>
      </span>
      <input id="dpi" type="number" min="100" max="25500" step="50" value="1600"/>
      <button id="apply-dpi" class="btn">Apply DPI</button>
    </div>
    <div class="row">
      <label for="rate">Report rate</label>
      <select id="rate">
        <option value="125">125 Hz</option>
        <option value="250">250 Hz</option>
        <option value="500">500 Hz</option>
        <option value="1000" selected>1000 Hz</option>
      </select>
      <button id="apply-rate" class="btn">Apply rate</button>
    </div>
  </section>

  <section class="card">
    <div class="row spread">
      <h2>Device info</h2>
      <button id="info" class="btn">Load info</button>
    </div>
    <pre id="info-out" class="output">—</pre>
  </section>

  <section class="card">
    <h2>Log</h2>
    <pre id="log" class="output">Ready.</pre>
  </section>
`;

const $ = (id) => document.getElementById(id);

function log(message) {
    const el = $('log');
    const stamp = new Date().toLocaleTimeString();
    el.textContent = `[${stamp}] ${message}\n${el.textContent}`;
}

function selectedDevice() {
    const value = $('device').value;
    return value || '';
}

async function refreshDevices() {
    try {
        const devices = await ListDevices();
        const select = $('device');
        select.innerHTML = '';
        if (!devices || devices.length === 0) {
            $('status').textContent = 'No libratbag devices found.';
            return;
        }
        for (const d of devices) {
            const option = document.createElement('option');
            option.value = d.name;
            option.textContent = `${d.name} (${d.codename})`;
            select.appendChild(option);
        }
        $('status').textContent = `${devices.length} device(s) found.`;
        log(`Found ${devices.length} device(s).`);
    } catch (err) {
        $('status').textContent = String(err);
        log(`ERROR: ${err}`);
    }
}

async function run(label, fn) {
    try {
        const out = await fn();
        log(`${label} -> ${out.trim() || 'ok'}`);
        await loadInfo();
    } catch (err) {
        log(`ERROR: ${label} failed: ${err}`);
    }
}

async function loadInfo() {
    const device = selectedDevice();
    if (!device) return;
    try {
        $('info-out').textContent = await GetDeviceInfo(device);
    } catch (err) {
        $('info-out').textContent = String(err);
    }
}

$('refresh').addEventListener('click', refreshDevices);
$('info').addEventListener('click', loadInfo);

$('activate').addEventListener('click', () =>
    run('Set active profile', () => SetActiveProfile(selectedDevice(), Number($('profile').value))));

$('apply-dpi').addEventListener('click', () =>
    run('Set DPI', () => SetDPI(
        selectedDevice(),
        Number($('profile').value),
        Number($('slot').value),
        Number($('dpi').value),
    )));

$('apply-rate').addEventListener('click', () =>
    run('Set report rate', () => SetReportRate(
        selectedDevice(),
        Number($('profile').value),
        Number($('rate').value),
    )));

(async () => {
    if (!(await RatbagctlAvailable())) {
        $('status').textContent = 'ratbagctl not found. Install libratbag and try again.';
        log('ERROR: ratbagctl is not installed.');
        return;
    }
    await refreshDevices();
    await loadInfo();
})();
