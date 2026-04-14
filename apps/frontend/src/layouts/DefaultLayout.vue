<script setup lang="ts">
import PublicTopbar from '@/components/layout/PublicTopbar.vue'
</script>

<template>
  <div class="layout-shell">
    <PublicTopbar />

    <main class="layout-main default-layout__main">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.default-layout__main {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem 1rem;
}

@media (min-width: 768px) {
  .default-layout__main {
    padding: 2.5rem 1.5rem;
  }
}
</style>
