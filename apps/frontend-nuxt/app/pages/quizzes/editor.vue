<script setup lang="ts">
definePageMeta({
  middleware: 'auth',
})

useHead({
  title: 'Редактор квиза',
})
</script>

<template>
  <section class="page-stub">Quiz editor page (next step)</section>
</template>

<style scoped>
.page-stub {
  display: grid;
  min-height: 100vh;
  place-content: center;
}
</style>
