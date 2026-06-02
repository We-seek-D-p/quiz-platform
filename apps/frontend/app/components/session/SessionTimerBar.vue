<script setup lang="ts">
const props = defineProps<{
  label: string
  progress: number
}>()

const clampProgress = (value: number) => Math.max(0, Math.min(100, value))

const displayedProgress = ref(clampProgress(props.progress))
const shouldAnimateDecrease = ref(true)

const normalizedProgress = computed(() => clampProgress(props.progress))
const fillStyle = computed(() => ({
  transform: `scaleX(${displayedProgress.value / 100})`,
}))

watch(
  normalizedProgress,
  (nextProgress) => {
    if (nextProgress > displayedProgress.value) {
      shouldAnimateDecrease.value = false
      displayedProgress.value = nextProgress

      if (import.meta.client) {
        requestAnimationFrame(() => {
          shouldAnimateDecrease.value = true
        })
      }
      return
    }

    shouldAnimateDecrease.value = true
    displayedProgress.value = nextProgress
  },
  { immediate: true },
)
</script>

<template>
  <div class="session-timer">
    <span class="session-timer__label">{{ label }}</span>
    <div
      class="session-timer__bar"
      role="progressbar"
      :aria-valuenow="Math.round(displayedProgress)"
      aria-valuemin="0"
      aria-valuemax="100"
      :aria-label="label"
    >
      <div
        class="session-timer__fill"
        :class="{ 'session-timer__fill--smooth': shouldAnimateDecrease }"
        :style="fillStyle"
      />
    </div>
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
  overflow: hidden;
  border-radius: 999px;
  background: color-mix(in srgb, var(--app-color-primary) 18%, transparent);
}

.session-timer__fill {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  transform-origin: left center;
  background: var(--app-color-primary);
  transition: none;
}

.session-timer__fill--smooth {
  transition: transform 120ms linear;
}
</style>
