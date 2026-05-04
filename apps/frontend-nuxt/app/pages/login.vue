<script setup lang="ts">
definePageMeta({
  middleware: 'guest',
})

useHead({
  title: 'Вход',
})
</script>

<template>
  <section class="page-stub">Login page (next step)</section>
</template>

<style scoped>
.page-stub {
  display: grid;
  min-height: 100vh;
  place-content: center;
}
</style>
