<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import FloatLabel from 'primevue/floatlabel'
import InputMask from 'primevue/inputmask'
import InputText from 'primevue/inputtext'
import { useToast } from 'primevue/usetoast'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const PLAYER_TOKEN_STORAGE_KEY = 'quiz:player_token'
const PLAYER_ROOM_CODE_STORAGE_KEY = 'quiz:room_code'

const roomCode = ref(typeof route.query.room_code === 'string' ? route.query.room_code.replace(/\D/g, '') : '')
const nickname = ref('')
const hasReconnectContext = ref(false)

const isJoinDisabled = computed(() => {
  return roomCode.value.trim().length !== 8 || nickname.value.trim().length < 2
})

const handleJoin = async () => {
  const normalizedCode = roomCode.value.replace(/\D/g, '').trim()
  const normalizedNickname = nickname.value.trim()

  if (normalizedCode.length !== 8) {
    toast.add({
      group: 'global',
      severity: 'warn',
      summary: 'Некорректный код',
      detail: 'Код комнаты должен содержать 8 цифр',
      life: 3000,
    })
    return
  }

  if (normalizedNickname.length < 2) {
    toast.add({
      group: 'global',
      severity: 'warn',
      summary: 'Некорректный ник',
      detail: 'Ник должен быть не короче 2 символов',
      life: 3000,
    })
    return
  }

  await router.push({
    path: '/game',
    query: {
      room_code: normalizedCode,
      nickname: normalizedNickname,
    },
  })
}

const continueLastGame = async () => {
  const storedRoomCode = import.meta.client ? (localStorage.getItem(PLAYER_ROOM_CODE_STORAGE_KEY) ?? '') : ''
  const normalizedCode = storedRoomCode.replace(/\D/g, '').trim()

  if (!normalizedCode || normalizedCode.length !== 8) {
    toast.add({
      group: 'global',
      severity: 'warn',
      summary: 'Нет данных для возврата',
      detail: 'Не найдена предыдущая игровая сессия',
      life: 3000,
    })
    return
  }

  await router.push({
    path: '/game',
    query: {
      room_code: normalizedCode,
    },
  })
}

onMounted(() => {
  if (!import.meta.client) {
    return
  }

  const storedToken = localStorage.getItem(PLAYER_TOKEN_STORAGE_KEY)
  const storedRoomCode = localStorage.getItem(PLAYER_ROOM_CODE_STORAGE_KEY)
  hasReconnectContext.value = Boolean(storedToken && storedRoomCode)
})

useHead({
  title: 'Вход в игру',
})
</script>

<template>
  <section class="mx-auto flex w-full max-w-(--app-card-narrow) flex-col gap-3">
    <Card class="w-full">
      <template #title>
        <div class="mb-2 text-center">Войти в игру</div>
      </template>

      <template #content>
        <div class="flex flex-col gap-3">
          <FloatLabel variant="in">
            <InputMask id="room_code" v-model="roomCode" mask="99999999" slot-char="" class="w-full" />
            <label for="room_code">Код комнаты</label>
          </FloatLabel>

          <FloatLabel variant="in">
            <InputText id="nickname" v-model="nickname" maxlength="32" class="w-full" />
            <label for="nickname">Ваш ник</label>
          </FloatLabel>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-center pt-3">
          <Button label="Подключиться" :disabled="isJoinDisabled" @click="handleJoin" />
        </div>
        <div v-if="hasReconnectContext" class="flex justify-center pt-3">
          <Button label="Продолжить прошлую игру" text icon="pi pi-history" @click="continueLastGame" />
        </div>
      </template>
    </Card>
  </section>
</template>
