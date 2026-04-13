<script setup lang="ts">
import Button from 'primevue/button'
import ThemeSwitcher from '@/components/common/ThemeSwitcher.vue'

withDefaults(
  defineProps<{
    title?: string
  }>(),
  {
    title: '',
  },
)

const emit = defineEmits<{
  menuClick: []
}>()
</script>

<template>
  <header
    class="flex h-14 items-center justify-between gap-4 border-b border-[var(--app-color-border)] bg-[var(--app-color-surface)] px-4 md:px-6"
  >
    <div class="flex min-w-0 items-center gap-3">
      <Button icon="pi pi-bars" text class="inline-flex" @click="emit('menuClick')" />
    </div>

    <div class="flex min-w-0 flex-1 justify-center">
      <p v-if="title" class="truncate text-sm font-semibold text-[var(--app-color-text-muted)]">
        {{ title }}
      </p>
    </div>

    <div class="flex items-center justify-end gap-3">
      <ThemeSwitcher />
    </div>
  </header>
</template>
