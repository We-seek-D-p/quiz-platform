<script setup lang="ts">
import Message from 'primevue/message'

defineProps<{
  error?: string | null
  reconnect?: string | null
}>()
</script>

<template>
  <div class="session-notice-stack">
    <Message v-if="error" severity="warn" :closable="false">{{ error }}</Message>
    <Message v-if="reconnect" severity="success" :closable="false">{{ reconnect }}</Message>
  </div>
</template>

<style scoped>
.session-notice-stack {
  display: grid;
  gap: 0.5rem;
}
</style>
