<script setup lang="ts">
import Toast from 'primevue/toast'
</script>

<template>
  <div class="app-root">
    <NuxtRouteAnnouncer />
    <NuxtLoadingIndicator />
    <Toast group="global" position="top-right" />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>
