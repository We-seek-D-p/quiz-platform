<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, type RouteLocationRaw } from 'vue-router'

const props = withDefaults(
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

const itemClass = computed(() => {
  const classes = [
    'flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left text-sm font-medium transition-colors',
  ]

  if (props.disabled) {
    classes.push('cursor-not-allowed opacity-60 text-[var(--app-color-text-muted)]')
    return classes
  }

  if (props.active) {
    classes.push('bg-[var(--app-color-surface-hover)] text-[var(--app-color-text-strong)]')
    return classes
  }

  classes.push(
    'text-[var(--app-color-text-muted)] hover:bg-[var(--app-color-surface-hover)] hover:text-[var(--app-color-text-strong)]',
  )

  return classes
})

const handleButtonClick = () => {
  if (props.disabled) {
    return
  }

  emit('select')
}

const handleLinkClick = (navigate: (event?: MouseEvent) => void, event: MouseEvent) => {
  if (props.disabled) {
    event.preventDefault()
    return
  }

  emit('select')
  navigate(event)
}
</script>

<template>
  <RouterLink v-if="to && !disabled" :to="to" custom v-slot="{ href, navigate }">
    <a
      :href="href"
      :class="itemClass"
      :aria-current="active ? 'page' : undefined"
      @click="handleLinkClick(navigate, $event)"
    >
      <i :class="[icon, 'text-base']" aria-hidden="true"></i>
      <span class="flex-1">{{ label }}</span>
    </a>
  </RouterLink>

  <button
    v-else
    type="button"
    :class="itemClass"
    :disabled="disabled"
    :aria-disabled="disabled"
    @click="handleButtonClick"
  >
    <i :class="[icon, 'text-base']" aria-hidden="true"></i>
    <span class="flex-1">{{ label }}</span>
  </button>
</template>
