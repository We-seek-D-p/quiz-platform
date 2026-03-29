<script setup lang="ts">
import { useRouter } from 'vue-router'

const props = withDefaults(
  defineProps<{
    size?: 'md' | 'lg'
  }>(),
  {
    size: 'md',
  },
)

const router = useRouter()

const goHome = () => {
  router.push('/')
}
</script>

<template>
  <button
    type="button"
    class="bg-transparent p-0 border-none text-left font-semibold text-[var(--app-color-primary)] transition-colors cursor-pointer select-none hover:text-[var(--app-color-primary-hover)]"
    :class="props.size === 'lg' ? 'text-xl' : 'text-lg'"
    @click="goHome"
  >
    Quiz Platform
  </button>
</template>
