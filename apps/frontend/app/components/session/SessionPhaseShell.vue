<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    subtitle?: string
    centered?: boolean
  }>(),
  {
    subtitle: '',
    centered: false,
  },
)
</script>

<template>
  <div class="flex flex-col gap-4" :class="{ 'items-center text-center': centered }">
    <h1 class="m-0 text-[clamp(1.5rem,2.6vw,2.2rem)] leading-[1.2]">{{ title }}</h1>
    <p v-if="subtitle" class="m-0 text-(--app-color-text-muted)">{{ subtitle }}</p>

    <slot />
  </div>
</template>
