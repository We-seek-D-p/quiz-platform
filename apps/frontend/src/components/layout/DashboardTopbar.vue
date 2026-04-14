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
  <header class="dashboard-topbar">
    <div class="dashboard-topbar__left">
      <Button icon="pi pi-bars" text @click="emit('menuClick')" />
    </div>

    <div class="dashboard-topbar__center">
      <p v-if="title" class="dashboard-topbar__title">
        {{ title }}
      </p>
    </div>

    <div class="dashboard-topbar__right">
      <ThemeSwitcher />
    </div>
  </header>
</template>

<style scoped>
.dashboard-topbar {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 1rem;
  height: var(--app-topbar-height);
  padding: 0 1rem;
  border-bottom: 1px solid var(--app-color-border);
  background-color: var(--app-color-surface);
}

.dashboard-topbar__left,
.dashboard-topbar__right {
  display: flex;
  align-items: center;
}

.dashboard-topbar__center {
  min-width: 0;
}

.dashboard-topbar__title {
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: center;
  color: var(--app-color-text-muted);
  font-size: 0.875rem;
  font-weight: 600;
}

@media (min-width: 768px) {
  .dashboard-topbar {
    padding: 0 1.5rem;
  }
}
</style>
