<script setup lang="ts">
definePageMeta({
  middleware: 'auth',
  layout: 'dashboard',
  title: 'Панель управления',
})

useHead({
  title: 'Панель управления',
})
</script>

<template>
  <section class="page-stub">Host dashboard page (next step)</section>
</template>

<style scoped>
.page-stub {
  display: grid;
  min-height: 100vh;
  place-content: center;
}
</style>
