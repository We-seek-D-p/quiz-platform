<script setup lang="ts">
defineProps<{
  icon: string
  label: string
}>()
</script>

<template>
  <button
    type="button"
    class="flex w-full items-center gap-3 px-3 py-2 rounded-lg text-left text-sm text-[var(--color-text-muted)] font-medium transition-colors hover:bg-[var(--color-surface-200)] hover:text-[var(--color-text-strong)] dark:text-[var(--color-text)] dark:hover:bg-[var(--color-surface-800)] dark:hover:text-[var(--color-text-inverse)]"
  >
    <i :class="[icon, 'text-base']" aria-hidden="true"></i>
    <span class="flex-1">{{ label }}</span>
  </button>
</template>
