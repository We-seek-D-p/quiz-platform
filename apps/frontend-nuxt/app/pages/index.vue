<script setup lang="ts">
useHead({
  title: 'Quiz Platform',
})
</script>

<template>
  <section class="home-stub">
    <h1>Quiz Platform Nuxt</h1>
  </section>
</template>

<style scoped>
.home-stub {
  display: grid;
  min-height: 100vh;
  place-content: center;
  gap: 0.5rem;
  padding: 1.5rem;
  text-align: center;
}

h1,
p {
  margin: 0;
}
</style>
