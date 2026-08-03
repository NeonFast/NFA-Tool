import './style.css';
import {
  GetVersion,
  ListAccounts,
  LoginFromKey,
  LoginSaved,
  DeleteAccount,
  ResetSteam,
  WindowMinimise,
  WindowClose,
} from '../wailsjs/go/main/App.js';

const $ = (id) => document.getElementById(id);

const statusBar = document.querySelector('.statusbar');
const statusEl = $('status');
const listEl = $('accounts-list');
const emptyEl = $('accounts-empty');
const keyInput = $('account-key');
const keepExisting = $('keep-existing');
const loginBtn = $('btn-login');

function setStatus(msg, kind = '') {
  statusEl.textContent = msg;
  statusBar.classList.remove('ok', 'err');
  if (kind) statusBar.classList.add(kind);
}

function escapeHtml(s) {
  return String(s)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;');
}

async function refreshAccounts() {
  try {
    const accounts = await ListAccounts();
    listEl.innerHTML = '';
    if (!accounts || accounts.length === 0) {
      emptyEl.classList.remove('hidden');
      return;
    }
    emptyEl.classList.add('hidden');
    accounts
      .slice()
      .sort((a, b) => a.name.localeCompare(b.name))
      .forEach((acc) => {
        const row = document.createElement('div');
        row.className = 'account-row';
        const expClass = acc.valid ? 'ok' : 'bad';
        row.innerHTML = `
          <div class="account-meta">
            <div class="name">${escapeHtml(acc.name)}</div>
            <div class="exp ${expClass}">${escapeHtml(acc.expiresIn)}</div>
          </div>
          <div class="row-actions">
            <button class="mini login" data-act="login">Login</button>
            <button class="mini del" data-act="del">Delete</button>
          </div>`;
        row.querySelector('[data-act="login"]').onclick = async () => {
          setStatus(`Logging in as ${acc.name}...`);
          const res = await LoginSaved(acc.name, keepExisting.checked);
          setStatus(res.message, res.ok ? 'ok' : 'err');
          await refreshAccounts();
        };
        row.querySelector('[data-act="del"]').onclick = async () => {
          const res = await DeleteAccount(acc.name);
          setStatus(res.message, res.ok ? 'ok' : 'err');
          await refreshAccounts();
        };
        listEl.appendChild(row);
      });
  } catch (e) {
    setStatus(String(e), 'err');
  }
}

async function doLogin() {
  const key = keyInput.value.trim();
  if (!key) {
    setStatus('Enter an account key', 'err');
    return;
  }
  loginBtn.disabled = true;
  setStatus('Logging in...');
  try {
    const res = await LoginFromKey(key, keepExisting.checked);
    setStatus(res.message, res.ok ? 'ok' : 'err');
    if (res.ok) {
      keyInput.value = '';
      await refreshAccounts();
    }
  } catch (e) {
    setStatus(String(e), 'err');
  } finally {
    loginBtn.disabled = false;
  }
}

async function init() {
  try {
    const v = await GetVersion();
    $('app-title').textContent = `NFA Tool v${v}`;
  } catch {
    $('app-title').textContent = 'NFA Tool';
  }

  $('btn-min').onclick = () => WindowMinimise();
  $('btn-close').onclick = () => WindowClose();
  $('btn-login').onclick = doLogin;
  keyInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') doLogin();
  });

  $('btn-reset').onclick = async () => {
    setStatus('Resetting Steam...');
    const res = await ResetSteam();
    setStatus(res.message, res.ok ? 'ok' : 'err');
  };

  const modal = $('modal');
  $('btn-help').onclick = () => modal.classList.remove('hidden');
  $('modal-close').onclick = () => modal.classList.add('hidden');
  modal.addEventListener('click', (e) => {
    if (e.target === modal) modal.classList.add('hidden');
  });

  await refreshAccounts();
  setStatus('Ready', 'ok');
}

init();
