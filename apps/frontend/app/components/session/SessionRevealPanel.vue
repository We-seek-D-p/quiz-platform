<script setup lang="ts">
import SessionLeaderboard from '~/components/session/SessionLeaderboard.vue'
import SessionPhaseShell from '~/components/session/SessionPhaseShell.vue'
import type { LeaderboardEntryView, SessionPhase } from '~/types/session-ws'

const props = defineProps<{
  phase: SessionPhase
  entries: LeaderboardEntryView[]
  score?: number | null
  rank?: number | null
}>()

const title = computed(() => {
  if (props.phase === 'answer_reveal') {
    return 'Промежуточный итог'
  }

  if (props.phase === 'finished') {
    return 'Игра завершена'
  }

  return 'Рейтинг игроков'
})

const subtitle = computed(() => {
  if (props.phase === 'answer_reveal') {
    return 'Время раскрытия ответов'
  }

  if (props.phase === 'finished') {
    return 'Финальный рейтинг'
  }

  return 'Готовьтесь к следующему раунду'
})
const hasPlayerScore = computed(() => props.score !== undefined && props.score !== null)
</script>

<template>
  <SessionPhaseShell :title="title" :subtitle="subtitle">
    <p v-if="hasPlayerScore" class="m-0 text-(--app-color-text-muted)">
      Ваш счет: {{ score }} · Место: {{ rank ?? '-' }}
    </p>

    <SessionLeaderboard :entries="entries" />

    <slot />
  </SessionPhaseShell>
</template>
