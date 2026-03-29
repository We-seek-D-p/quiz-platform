<script setup lang="ts">
defineProps<{
  icon: string
  label: string
}>()
</script>

<template>
  <button
    type="button"
    class="flex w-full items-center gap-3 px-3 py-2 rounded-lg text-left text-sm font-medium text-[var(--app-color-text-muted)] transition-colors hover:bg-[var(--app-color-surface-hover)] hover:text-[var(--app-color-text-strong)]"
  >
    <i :class="[icon, 'text-base']" aria-hidden="true"></i>
    <span class="flex-1">{{ label }}</span>
  </button>
</template>
