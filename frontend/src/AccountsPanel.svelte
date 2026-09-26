<script lang="ts">
  import { t, localizeExpiry } from './i18n';
  import {
    app,
    selectedNames,
    allSelected,
    toggleSelectAll,
    exportAccounts,
    deleteSelected,
    paintSelect,
    startDragSelect,
    endDragSelect,
    exportOne,
    openAccInfo,
    loginSaved,
    deleteAccount,
    toggleShowPersona,
    openBulkView,
    type Account,
  } from './state.svelte.js';

  let accQuery = $state('');

  const filteredAccounts = $derived(
    accQuery.trim()
      ? app.accounts.filter((a) => {
          const q = accQuery.trim().toLowerCase();
          return (
            a.name.toLowerCase().includes(q) ||
            (a.persona ?? '').toLowerCase().includes(q) ||
            (a.steamId ?? '').includes(q)
          );
        })
      : app.accounts,
  );

  function accDisplay(acc: Account): string {
    return app.showPersona && acc.persona ? acc.persona : acc.name;
  }
</script>

<section class="panel accounts-panel">
  <div class="accounts-head">
    <div>
      <div class="panel-label"><span class="lbl-t">{t(app.lang, 'savedAccounts')}</span></div>
      {#if app.accounts.length > 0}
        {#key app.accounts.length}
          <div class="count-chip">{app.accounts.length}</div>
        {/key}
      {/if}
    </div>
    <div class="head-actions">
      {#if app.accounts.length > 0}
        <button
          class="btn ghost sm"
          type="button"
          title={t(app.lang, 'viewPersonaHint')}
          onclick={toggleShowPersona}
        >
          {app.showPersona ? t(app.lang, 'viewPersonas') : t(app.lang, 'viewLogins')}
        </button>
      {/if}
      {#if app.mode === 'advanced'}
        <button class="btn ghost sm" type="button" onclick={openBulkView}>
          {t(app.lang, 'bulkCheckBtn')}
        </button>
      {/if}
      {#if app.accounts.length > 0 && app.mode === 'advanced'}
      <div class="export-bar">
        <button class="btn ghost sm" type="button" disabled={app.exportBusy} onclick={toggleSelectAll}>
          {allSelected() ? t(app.lang, 'deselectAll') : t(app.lang, 'selectAll')}
        </button>
        {#if selectedNames().length > 0}
          <button
            class="btn ghost sm"
            type="button"
            disabled={app.exportBusy}
            onclick={() => exportAccounts(false)}
          >
            {t(app.lang, 'exportSelected')}
            {#key selectedNames().length}
              <span class="num-pop">({selectedNames().length})</span>
            {/key}
          </button>
        {/if}
        <button class="btn ghost sm" type="button" disabled={app.exportBusy} onclick={() => exportAccounts(true)}>
          {t(app.lang, 'exportAll')}
        </button>
        {#if selectedNames().length > 0}
          <button
            class="btn ghost sm danger-sm"
            type="button"
            onclick={deleteSelected}
          >
            {t(app.lang, 'deleteSelected')}
            {#key selectedNames().length}
              <span class="num-pop">({selectedNames().length})</span>
            {/key}
          </button>
        {/if}
      </div>
      {/if}
    </div>
  </div>
  {#if app.accounts.length > 0}
    <label class="field acc-search">
      <input
        type="text"
        spellcheck="false"
        autocomplete="off"
        placeholder={t(app.lang, 'searchAccounts')}
        bind:value={accQuery}
      />
    </label>
  {/if}
  {#if app.accounts.length === 0}
    <div class="empty">
      <div class="empty-icon" aria-hidden="true">∅</div>
      <p>{t(app.lang, 'noSavedAccounts')}</p>
    </div>
  {:else if filteredAccounts.length === 0}
    <div class="empty">
      <p>{t(app.lang, 'searchEmpty')}</p>
    </div>
  {:else}
    <div class="accounts-list" class:painting={app.dragSelect}>
      {#each filteredAccounts as acc, i (acc.name)}
        <div
          class="account-row"
          class:simple={app.mode === 'simple'}
          class:picked={app.mode === 'advanced' && !!app.selected[acc.name]}
          data-acc={acc.name}
          style="animation-delay: {Math.min(i * 30, 300)}ms"
          onpointerenter={() => paintSelect(acc.name)}
          onpointerdown={(e) => {
            const t = e.target as HTMLElement;
            if (t.closest('.row-actions')) return;
            startDragSelect(e, acc.name);
          }}
        >
          {#if app.mode === 'advanced'}
            <label class="pick">
              <input type="checkbox" checked={!!app.selected[acc.name]} tabindex="-1" />
              <span class="box"></span>
            </label>
          {/if}
          {#if acc.avatar}
            <img class="acc-ava" src={acc.avatar} alt="" />
          {:else}
            <div class="acc-ava placeholder" aria-hidden="true">{accDisplay(acc).slice(0, 1).toUpperCase()}</div>
          {/if}
          <div class="meta">
            <div class="name" title={acc.name}>{accDisplay(acc)}</div>
            <div class="exp" class:ok={acc.valid} class:bad={!acc.valid}>
              {localizeExpiry(app.lang, acc.expiresIn)}
            </div>
          </div>
          <div
            class="row-actions"
            onpointerdown={(e) => {
              e.stopPropagation();
              endDragSelect();
            }}
          >
            {#if app.mode === 'advanced'}
              <button
                class="btn mini export"
                type="button"
                disabled={app.exportBusy}
                onclick={(e) => {
                  e.stopPropagation();
                  exportOne(acc.name);
                }}
              >{t(app.lang, 'export')}</button>
            {/if}
            <button
              class="btn mini info"
              type="button"
              onclick={(e) => {
                e.stopPropagation();
                openAccInfo(acc.name);
              }}
            >{t(app.lang, 'info')}</button>
            <button
              class="btn mini login"
              type="button"
              onclick={(e) => {
                e.stopPropagation();
                loginSaved(acc.name);
              }}
            >{t(app.lang, 'login')}</button>
            <button
              class="btn mini del"
              type="button"
              onclick={(e) => deleteAccount(acc.name, e)}
            >{t(app.lang, 'delete')}</button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</section>
