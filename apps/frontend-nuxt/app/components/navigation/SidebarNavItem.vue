<script setup lang="ts">
type Props = {
  icon: string
  label: string
  to: string
}

defineProps<Props>()
</script>

<template>
  <NuxtLink :to="to" class="sidebar-nav-item" active-class="sidebar-nav-item--active">
    <i :class="[icon, 'sidebar-nav-item__icon']" aria-hidden="true"></i>
    <span>{{ label }}</span>
  </NuxtLink>
</template>

<style scoped>
.sidebar-nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 0.75rem;
  border-radius: var(--app-control-radius);
  color: var(--app-color-text-muted);
  font-size: 0.875rem;
  font-weight: 500;
  text-decoration: none;
  transition:
    background-color var(--app-transition-fast),
    color var(--app-transition-fast);
}

.sidebar-nav-item:hover,
.sidebar-nav-item--active {
  background-color: var(--app-color-surface-hover);
  color: var(--app-color-text-strong);
}

.sidebar-nav-item__icon {
  font-size: 1rem;
}
</style>
