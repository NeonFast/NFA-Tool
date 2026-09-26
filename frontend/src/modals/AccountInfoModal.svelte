<script lang="ts">
  import { AppService } from '../../bindings/nfa-tool';
  import { t } from '../i18n';
  import {
    app,
    trOnline,
    trVisibility,
    yn,
    checkDesc,
    checkLabel,
    loginFromInfo,
    saveFromInfo,
  } from '../state.svelte.js';
</script>

<div class="modal stacked" role="dialog" aria-modal="true" tabindex="-1">
  <div class="modal-card update">
    <div class="modal-head">
      <h3>{t(app.lang, 'accInfoTitle')}</h3>
      <button class="modal-x" type="button" onclick={() => (app.showAccInfo = false)} aria-label={t(app.lang, 'close')}>✕</button>
    </div>
    {#if app.infoBusy}
      <p class="update-msg wait">{t(app.lang, 'accInfoLoading')}</p>
    {:else if app.infoData}
      <div class="accinfo-head">
        {#if app.infoData.avatarFull}
          <img class="accinfo-ava" src={app.infoData.avatarFull} alt="" />
        {/if}
        <div class="accinfo-id-block">
          <div class="accinfo-nick">{app.infoData.personaName || '—'}</div>
          <div class="accinfo-id">{app.infoData.steamId}</div>
        </div>
      </div>
      <div class="accinfo-grid">
        <span class="ai-label">{t(app.lang, 'rowStatus')}</span>
        <span class="ai-value">{app.infoData.profileErr ? '—' : trOnline(app.infoData.onlineState)}</span>
        <span class="ai-label">{t(app.lang, 'rowVisibility')}</span>
        <span class="ai-value">{app.infoData.profileErr ? '—' : trVisibility(app.infoData.visibility)}</span>
        {#if app.infoData.realName}
          <span class="ai-label">{t(app.lang, 'rowRealName')}</span>
          <span class="ai-value">{app.infoData.realName}</span>
        {/if}
        {#if app.infoData.location}
          <span class="ai-label">{t(app.lang, 'rowLocation')}</span>
          <span class="ai-value">{app.infoData.location}</span>
        {/if}
        {#if app.infoData.inGame}
          <span class="ai-label">{t(app.lang, 'rowInGame')}</span>
          <span class="ai-value good">{app.infoData.inGame}</span>
        {/if}
        {#if app.infoData.level}
          <span class="ai-label">{t(app.lang, 'rowLevel')}</span>
          <span class="ai-value">{app.infoData.level}</span>
        {/if}
        {#if app.infoData.gamesCount}
          <span class="ai-label">{t(app.lang, 'rowGames')}</span>
          <span class="ai-value">{app.infoData.gamesCount}</span>
        {/if}
        {#if app.infoData.friendsCount > 0}
          <span class="ai-label">{t(app.lang, 'rowFriends')}</span>
          <span class="ai-value">{app.infoData.friendsCount}</span>
        {/if}
        {#if app.infoData.topGame}
          <span class="ai-label">{t(app.lang, 'rowTopGame')}</span>
          <span class="ai-value">{app.infoData.topGame}{app.infoData.topGameHours ? ` · ${app.infoData.topGameHours}h` : ''}</span>
        {/if}
        <span class="ai-label">{t(app.lang, 'rowVac')}</span>
        <span class="ai-value" class:bad={app.infoData.vacBanned}>{app.infoData.profileErr ? '—' : yn(app.infoData.vacBanned)}</span>
        {#if app.infoData.banInfo}
          <span class="ai-label">{t(app.lang, 'rowBanInfo')}</span>
          <span class="ai-value bad">{app.infoData.banInfo}{app.infoData.banDays ? ` · ${app.infoData.banDays} ${t(app.lang, 'daysShort')}` : ''}</span>
        {/if}
        {#if app.infoData.cs2Items > 0}
          <span class="ai-label">{t(app.lang, 'rowCS2Items')}</span>
          <span class="ai-value">{app.infoData.cs2Items}{app.infoData.cs2Rarity ? ` · ${app.infoData.cs2Rarity}` : ''}</span>
        {/if}
        {#if app.infoData.cs2Hours}
          <span class="ai-label">{t(app.lang, 'rowCS2Hours')}</span>
          <span class="ai-value">{app.infoData.cs2Hours}h</span>
        {/if}
        {#if app.infoData.cs2Medals}
          <span class="ai-label">{t(app.lang, 'rowCS2Medals')}</span>
          <span class="ai-value" title={app.infoData.cs2Medals}>{app.infoData.cs2Medals}</span>
        {/if}
        {#if app.infoData.inventoryValue}
          <span class="ai-label">{t(app.lang, 'rowInvValue')}</span>
          <span class="ai-value good">
            {app.infoData.inventoryValue}{app.infoData.inventoryPartial ? ` ${t(app.lang, 'invPartial')}` : ''}
          </span>
        {/if}
        {#if app.infoData.cs2Prime}
          <span class="ai-label">{t(app.lang, 'rowPrime')}</span>
          <span class="ai-value" class:good={app.infoData.cs2Prime === 'yes'}>{yn(app.infoData.cs2Prime === 'yes')}</span>
        {/if}
        {#if app.infoData.licensesCount > 0}
          <span class="ai-label">{t(app.lang, 'rowLicenses')}</span>
          <span class="ai-value">{app.infoData.licensesCount}</span>
        {/if}
        <span class="ai-label">{t(app.lang, 'rowTrade')}</span>
        <span class="ai-value" class:bad={!!app.infoData.tradeBan && app.infoData.tradeBan.toLowerCase() !== 'none'}>
          {app.infoData.profileErr ? '—' : !app.infoData.tradeBan || app.infoData.tradeBan.toLowerCase() === 'none' ? t(app.lang, 'no') : app.infoData.tradeBan}
        </span>
        <span class="ai-label">{t(app.lang, 'rowLimited')}</span>
        <span class="ai-value" class:bad={app.infoData.limited}>{app.infoData.profileErr ? '—' : yn(app.infoData.limited)}</span>
        <span class="ai-label">{t(app.lang, 'rowSince')}</span>
        <span class="ai-value">{app.infoData.memberSince || '—'}</span>
        {#if app.infoData.email}
          <span class="ai-label">Email</span>
          <span class="ai-value">{app.infoData.email}</span>
        {/if}
        {#if app.infoData.wallet}
          <span class="ai-label">{t(app.lang, 'rowWallet')}</span>
          <span class="ai-value">{app.infoData.wallet}</span>
        {/if}
        {#if app.infoData.country}
          <span class="ai-label">{t(app.lang, 'rowCountry')}</span>
          <span class="ai-value">{app.infoData.country}</span>
        {/if}
        <span class="ai-label">{t(app.lang, 'rowToken')}</span>
        <span class="ai-value" class:good={app.infoData.tokenAlive} class:bad={!app.infoData.tokenAlive}>{yn(app.infoData.tokenAlive)}</span>
      </div>
      {#if app.infoData.tokenSteamId || app.infoData.tokenIssued || app.infoData.tokenExpires}
        <div class="panel-label tok-label">{t(app.lang, 'tokSection')}</div>
        <div class="accinfo-grid">
          {#if app.infoData.tokenSteamId}
            <span class="ai-label">{t(app.lang, 'rowTokSteamID')}</span>
            <span class="ai-value">{app.infoData.tokenSteamId}</span>
          {/if}
          {#if app.infoData.tokenIssued}
            <span class="ai-label">{t(app.lang, 'rowTokIssued')}</span>
            <span class="ai-value">{app.infoData.tokenIssued}</span>
          {/if}
          {#if app.infoData.tokenExpires}
            <span class="ai-label">{t(app.lang, 'rowTokExpires')}</span>
            <span class="ai-value" class:bad={!app.infoData.tokenDaysLeft}>
              {app.infoData.tokenExpires}{app.infoData.tokenDaysLeft ? ` · ${app.infoData.tokenDaysLeft} ${t(app.lang, 'daysShort')}` : ''}
            </span>
          {/if}
          {#if app.infoData.tokenAudiences}
            <span class="ai-label">{t(app.lang, 'rowTokAud')}</span>
            <span class="ai-value" title={app.infoData.tokenAudiences}>{app.infoData.tokenAudiences}</span>
          {/if}
        </div>
      {/if}
      {#if app.infoData.tokenChecks?.length}
        <div class="panel-label tok-label">{t(app.lang, 'checksTitle')}</div>
        <div class="accinfo-grid checks-grid">
          {#each app.infoData.tokenChecks as c (c.id)}
            <span class="ai-label" title={checkDesc(c.id)}>{checkLabel(c.id)}</span>
            <span
              class="ai-value"
              title={checkDesc(c.id)}
              class:good={c.status === 'ok'}
              class:bad={c.status === 'fail'}
            >{c.status === 'ok' ? '✓' : c.status === 'fail' ? '✗' : '—'}</span>
          {/each}
        </div>
      {/if}
      {#if app.infoData.tokenClaimsJson}
        <details class="claims">
          <summary>{t(app.lang, 'rowTokClaims')}</summary>
          <pre class="claims-pre">{app.infoData.tokenClaimsJson}</pre>
        </details>
      {/if}
      {#if app.infoData.summary}
        <div class="accinfo-summary">{app.infoData.summary}</div>
      {/if}
      {#if app.infoLogin}
        <button class="btn primary" type="button" disabled={app.loading} onclick={loginFromInfo}>
          {app.loading ? t(app.lang, 'working') : t(app.lang, 'login')}
        </button>
      {/if}
      {#if app.infoLogin?.kind === 'key'}
        <button class="btn ghost block" type="button" disabled={app.loading} onclick={saveFromInfo}>
          {t(app.lang, 'saveBtn')}
        </button>
      {/if}
      {#if app.infoData.profileUrl}
        <button
          class="btn ghost block"
          type="button"
          onclick={() => (AppService as any).OpenURL(app.infoData!.profileUrl)}
        >{t(app.lang, 'accInfoOpenProfile')}</button>
      {/if}
    {/if}
  </div>
</div>
