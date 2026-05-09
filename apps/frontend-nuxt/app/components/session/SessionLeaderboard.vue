<script setup lang="ts">
import type { LeaderboardEntryView } from '~/types/session-ws'

defineProps<{
  entries: LeaderboardEntryView[]
}>()
</script>

<template>
  <ol class="session-leaderboard" v-if="entries.length > 0">
    <li v-for="entry in entries" :key="`${entry.nickname}-${entry.rank}`">
      <span>{{ entry.rank }}. {{ entry.nickname }}</span>
      <strong>{{ entry.score }}</strong>
    </li>
  </ol>
</template>

<style scoped>
.session-leaderboard {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.session-leaderboard li {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border: 1px solid var(--app-color-border);
  border-radius: 0.75rem;
  padding: 0.65rem 0.75rem;
}
</style>
