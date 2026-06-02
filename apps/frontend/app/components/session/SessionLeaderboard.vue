<script setup lang="ts">
import type { LeaderboardEntryView } from '~/types/session-ws'

const props = defineProps<{
  entries: LeaderboardEntryView[]
}>()

const topEntries = computed(() => props.entries.slice(0, 10))
</script>

<template>
  <div v-if="topEntries.length > 0" class="session-leaderboard" role="table" aria-label="Топ игроков">
    <div class="session-leaderboard__row session-leaderboard__row--head" role="row">
      <span role="columnheader">#</span>
      <span role="columnheader">Игрок</span>
      <span role="columnheader">Очки</span>
    </div>

    <div v-for="entry in topEntries" :key="`${entry.nickname}-${entry.rank}`" class="session-leaderboard__row" role="row">
      <span role="cell">{{ entry.rank }}</span>
      <span role="cell">{{ entry.nickname }}</span>
      <strong role="cell">{{ entry.score }}</strong>
    </div>
  </div>
</template>

<style scoped>
.session-leaderboard {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: 100%;
}

.session-leaderboard__row {
  display: grid;
  grid-template-columns: 3rem minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.75rem;
  border: 1px solid var(--app-color-border);
  border-radius: var(--app-control-radius);
  padding: 0.65rem 0.75rem;
}

.session-leaderboard__row--head {
  border-color: transparent;
  padding-top: 0;
  padding-bottom: 0;
  color: var(--app-color-text-muted);
  font-size: 0.8rem;
  font-weight: 700;
}

.session-leaderboard__row span:nth-child(2) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
