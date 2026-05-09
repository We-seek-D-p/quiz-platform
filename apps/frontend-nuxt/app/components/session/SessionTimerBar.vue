<script setup lang="ts">
import ProgressBar from 'primevue/progressbar'

defineProps<{
  label: string
  progress: number
}>()
</script>

<template>
  <div class="session-timer">
    <span class="session-timer__label">{{ label }}</span>
    <ProgressBar :value="progress" :show-value="false" class="session-timer__bar" />
  </div>
</template>

<style scoped>
.session-timer {
  display: grid;
  gap: 0.5rem;
}

.session-timer__label {
  justify-self: end;
  font-weight: 700;
}

.session-timer__bar {
  height: 0.625rem;
}
</style>
