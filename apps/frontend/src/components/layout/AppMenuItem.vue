<script setup lang="ts">
defineProps<{
  icon: string
  label: string
}>()
</script>

<template>
  <button
    type="button"
    class="w-full flex items-center gap-3 rounded-md px-3 py-2 text-left text-surface-700 hover:bg-surface-100 dark:text-surface-200 dark:hover:bg-surface-800 transition-colors"
  >
    <i :class="icon" class="text-base" aria-hidden="true"></i>
    <span class="text-sm font-medium">{{ label }}</span>
  </button>
</template>
