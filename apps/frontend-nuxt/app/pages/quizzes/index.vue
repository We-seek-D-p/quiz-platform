<script setup lang="ts">
definePageMeta({
  middleware: 'auth',
  layout: 'dashboard',
  title: 'Мои квизы',
})

useHead({
  title: 'Мои квизы',
})
</script>

<template>
  <section class="page-stub">Quizzes list page (next step)</section>
</template>

<style scoped>
.page-stub {
  display: grid;
  min-height: 100vh;
  place-content: center;
}
</style>
