<script setup lang="ts">
import Tag from 'primevue/tag'
import type { ConnectionStatus } from '~/types/session-ws'

const props = defineProps<{
  status: ConnectionStatus
  roomCode?: string | null
  roomPrefix?: string
}>()

const isConnected = computed(() => props.status === 'connected')
const isReconnecting = computed(() => props.status === 'reconnecting')

const statusSeverity = computed(() => {
  return isConnected.value ? 'success' : isReconnecting.value ? 'warn' : 'danger'
})

const statusLabel = computed(() => {
  return isConnected.value ? 'Подключено' : isReconnecting.value ? 'Переподключение' : 'Не в сети'
})
</script>

<template>
  <div class="session-connection">
    <Tag :severity="statusSeverity" :value="statusLabel" />
    <span v-if="roomCode" class="session-connection__room">{{ roomPrefix ?? '' }}{{ roomCode }}</span>
  </div>
</template>

<style scoped>
.session-connection {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.session-connection__room {
  font-weight: 700;
}
</style>
