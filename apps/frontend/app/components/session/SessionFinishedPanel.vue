<script setup lang="ts">
import Button from 'primevue/button'
import SessionRevealPanel from '~/components/session/SessionRevealPanel.vue'
import type { LeaderboardEntryView } from '~/types/session-ws'

defineProps<{
  entries: LeaderboardEntryView[]
  rank?: number | null
}>()

const router = useRouter()
</script>

<template>
  <SessionRevealPanel phase="finished" :entries="entries">
    <p v-if="rank !== undefined && rank !== null" class="m-0 text-(--app-color-text-muted)">
      Ваше место: {{ rank }}
    </p>

    <div class="flex flex-wrap items-center gap-2">
      <Button label="Вернуться на главную" outlined @click="router.replace('/')" />
    </div>
  </SessionRevealPanel>
</template>
