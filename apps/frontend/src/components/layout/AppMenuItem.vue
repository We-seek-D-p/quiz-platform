<script setup lang="ts">
import { RouterLink, type RouteLocationRaw } from 'vue-router'

withDefaults(
  defineProps<{
    icon: string
    label: string
    to?: RouteLocationRaw
    active?: boolean
    disabled?: boolean
  }>(),
  {
    to: undefined,
    active: false,
    disabled: false,
  },
)

const emit = defineEmits<{
  select: []
}>()

const handleClick = () => {
  emit('select')
}
</script>

<template>
  <RouterLink
    v-if="to && !disabled"
    :to="to"
    class="app-menu-item"
    :class="{ 'app-menu-item--active': active }"
    :aria-current="active ? 'page' : undefined"
    @click="handleClick"
  >
    <i :class="[icon, 'app-menu-item__icon']" aria-hidden="true"></i>
    <span class="app-menu-item__label">{{ label }}</span>
  </RouterLink>

  <button
    v-else
    type="button"
    class="app-menu-item"
    :class="{
      'app-menu-item--active': active && !disabled,
      'app-menu-item--disabled': disabled,
    }"
    :disabled="disabled"
    :aria-disabled="disabled"
    @click="handleClick"
  >
    <i :class="[icon, 'app-menu-item__icon']" aria-hidden="true"></i>
    <span class="app-menu-item__label">{{ label }}</span>
  </button>
</template>

<style scoped>
.app-menu-item {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 0.75rem;
  border: none;
  border-radius: var(--app-control-radius);
  background: transparent;
  color: var(--app-color-text-muted);
  font: inherit;
  font-size: 0.875rem;
  font-weight: 500;
  text-align: left;
  text-decoration: none;
  cursor: pointer;
  transition:
    background-color var(--app-transition-fast),
    color var(--app-transition-fast),
    opacity var(--app-transition-fast);
}

.app-menu-item:hover {
  background-color: var(--app-color-surface-hover);
  color: var(--app-color-text-strong);
}

.app-menu-item--active {
  background-color: var(--app-color-surface-hover);
  color: var(--app-color-text-strong);
}

.app-menu-item--disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.app-menu-item--disabled:hover {
  background-color: transparent;
  color: var(--app-color-text-muted);
}

.app-menu-item__icon {
  font-size: 1rem;
}

.app-menu-item__label {
  flex: 1;
}
</style>
