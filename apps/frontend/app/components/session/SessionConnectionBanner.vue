<script setup lang="ts">
import type { ConnectionStatus } from '~/types/session-ws'

const props = defineProps<{
  status: ConnectionStatus
  roomCode?: string | null
  roomPrefix?: string
}>()

const isConnected = computed(() => props.status === 'connected')
const isReconnecting = computed(() => props.status === 'reconnecting')

const statusTone = computed(() => {
  if (isConnected.value) {
    return 'connected'
  }

  if (props.status === 'connecting' || isReconnecting.value) {
    return 'reconnecting'
  }

  return 'disconnected'
})

const statusLabel = computed(() => {
  return isConnected.value ? 'Подключено' : statusTone.value === 'reconnecting' ? 'Переподключение' : 'Не подключено'
})
</script>

<template>
  <div class="session-connection">
    <span class="session-connection__status" :class="`session-connection__status--${statusTone}`">
      <span class="session-connection__dot" />
      <span>{{ statusLabel }}</span>
    </span>
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

.session-connection__status {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-weight: 700;
}

.session-connection__dot {
  width: 0.625rem;
  height: 0.625rem;
  border-radius: 999px;
  background: currentColor;
  box-shadow: 0 0 0 0.1875rem color-mix(in srgb, currentColor 18%, transparent);
}

.session-connection__status--connected {
  color: var(--p-green-500);
}

.session-connection__status--reconnecting {
  color: var(--p-yellow-500);
}

.session-connection__status--disconnected {
  color: var(--p-red-500);
}

.session-connection__room {
  font-weight: 700;
}
</style>
