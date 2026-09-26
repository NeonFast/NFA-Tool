<script lang="ts">
  import { onMount } from 'svelte';
  import { fly } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { Events } from '@wailsio/runtime';
  import { AppService } from '../bindings/nfa-tool';
  import { installLinkGuard } from './actions';
  import { t } from './i18n';
  import {
    app,
    type BulkItem,
    refreshAccounts,
    checkUpdates,
    endDragSelect,
  } from './state.svelte.js';
  import Titlebar from './Titlebar.svelte';
  import SysPanel from './SysPanel.svelte';
  import LogsPanel from './LogsPanel.svelte';
  import LoginPanel from './LoginPanel.svelte';
  import AccountsPanel from './AccountsPanel.svelte';
  import Statusbar from './Statusbar.svelte';
  import ExportPickModal from './modals/ExportPickModal.svelte';
  import HintConfirmModal from './modals/HintConfirmModal.svelte';
  import AccountInfoModal from './modals/AccountInfoModal.svelte';
  import HelpModal from './modals/HelpModal.svelte';
  import SettingsModal from './modals/SettingsModal.svelte';
  import DeleteConfirmModal from './modals/DeleteConfirmModal.svelte';
  import BulkCheckPanel from './BulkCheckPanel.svelte';
  import ExportToast from './modals/ExportToast.svelte';
  import BusyModal from './modals/BusyModal.svelte';
  import SuccessModal from './modals/SuccessModal.svelte';
  import UpdateModal from './modals/UpdateModal.svelte';

  onMount(async () => {
    installLinkGuard();
    app.status = t(app.lang, 'ready');
    window.addEventListener('pointerup', endDragSelect);
    window.addEventListener('pointercancel', endDragSelect);
    window.addEventListener('blur', endDragSelect);
    const onJsError = (e: ErrorEvent) => {
      void (AppService as any).LogFrontendError?.(`${e.message} @ ${e.filename}:${e.lineno}`);
    };
    const onRejection = (e: PromiseRejectionEvent) => {
      void (AppService as any).LogFrontendError?.(`unhandledrejection: ${String(e.reason)}`);
    };
    window.addEventListener('error', onJsError);
    window.addEventListener('unhandledrejection', onRejection);
    const offStage = Events.On('login:stage', (ev) => {
      const st = (ev as { data?: { stage?: string } })?.data?.stage;
      if (typeof st === 'string') app.loginStage = st;
    });
    const offBulk = Events.On('bulk:item', (ev) => {
      if (!app.bulkBusy) return;
      const it = (ev as { data?: BulkItem })?.data;
      if (it) app.bulkLive = [...app.bulkLive, it];
    });
    const offAvatar = Events.On('avatar:ready', (ev) => {
      const d = (ev as { data?: { name?: string; steamId?: string; avatar?: string; persona?: string } })?.data;
      if (!d?.avatar && !d?.persona) return;
      app.accounts = app.accounts.map((a) =>
        (d.steamId && a.steamId === d.steamId) || (d.name && a.name === d.name)
          ? { ...a, avatar: d.avatar || a.avatar, persona: d.persona || a.persona }
          : a,
      );
    });
    let avatarSyncTimer: ReturnType<typeof setTimeout> | undefined;
    const offAvatarProg = Events.On('avatar:progress', (ev) => {
      const d = (ev as { data?: { done?: number; total?: number } })?.data;
      clearTimeout(avatarSyncTimer);
      if (!d || !d.total || (d.done ?? 0) >= d.total) {
        avatarSyncTimer = setTimeout(() => (app.avatarSync = null), 1200);
        return;
      }
      app.avatarSync = { done: d.done ?? 0, total: d.total };
    });
    try {
      const anySvc = AppService as typeof AppService & { GetAppName?: () => Promise<string> };
      if (typeof anySvc.GetAppName === 'function') {
        app.appName = await anySvc.GetAppName();
      }
      app.version = await AppService.GetVersion();
    } catch {
    }
    await refreshAccounts();
    void checkUpdates(true);
    return () => {
      window.removeEventListener('pointerup', endDragSelect);
      window.removeEventListener('pointercancel', endDragSelect);
      window.removeEventListener('blur', endDragSelect);
      window.removeEventListener('error', onJsError);
      window.removeEventListener('unhandledrejection', onRejection);
      offStage();
      offBulk();
      offAvatar();
      offAvatarProg();
    };
  });
</script>

<div class="app">
  <Titlebar />

  <div class="swap-wrap">
  {#key app.bulkView ? 'bulk' : app.mode}
    <main
      class="content"
      class:mode-swap={app.swapped}
      in:fly={{ x: app.swapped ? app.slideFrom : 0, duration: 260, easing: cubicOut }}
      out:fly={{ x: app.swapped ? -app.slideFrom : 0, duration: 220, easing: cubicOut }}
    >
      {#if app.bulkView}
        <BulkCheckPanel />
      {:else if app.mode === 'logs'}
        <SysPanel />
        <LogsPanel />
      {:else}
        <LoginPanel />
        <AccountsPanel />
      {/if}
    </main>
  {/key}
  </div>

  <Statusbar />
</div>

{#if app.showExportPick}
  <ExportPickModal />
{/if}

{#if app.showHintConfirm}
  <HintConfirmModal />
{/if}

{#if app.deleteConfirm}
  <DeleteConfirmModal />
{/if}

{#if app.showAccInfo}
  <AccountInfoModal />
{/if}

{#if app.showHelp}
  <HelpModal />
{/if}

{#if app.showSettings}
  <SettingsModal />
{/if}

{#if app.exportToast}
  <ExportToast />
{/if}

{#if app.busyOpen}
  <BusyModal />
{/if}

{#if app.successOpen}
  <SuccessModal />
{/if}

{#if app.showUpdate && app.updateInfo}
  <UpdateModal />
{/if}
