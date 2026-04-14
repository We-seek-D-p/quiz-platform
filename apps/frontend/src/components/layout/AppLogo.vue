<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <RouterLink to="/" class="app-logo"> Quiz Platform </RouterLink>
</template>

<style scoped>
.app-logo {
  color: var(--app-color-primary);
  font-size: 1.25rem;
  font-weight: 600;
  text-decoration: none;
  transition: color var(--app-transition-fast);
}

.app-logo:hover {
  color: var(--app-color-primary-hover);
}
</style>
