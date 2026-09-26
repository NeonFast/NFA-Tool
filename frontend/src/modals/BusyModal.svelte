<script lang="ts">
  import { app, busyProgress, stageLabel } from '../state.svelte.js';
</script>

<div class="modal busy-modal">
  <div class="modal-card busy-card" role="status">
    <div class="busy-ring" aria-hidden="true"></div>
    <div class="busy-title">{app.busyTitle}</div>
    {#if app.busySteps}
      <div class="busy-bar"><span style="width: {busyProgress()}%"></span></div>
      {#key app.loginStage}
        <div class="busy-stage">{stageLabel(app.loginStage)}</div>
      {/key}
    {:else if app.busyMsg}
      <div class="busy-stage">{app.busyMsg}</div>
    {/if}
  </div>
</div>
