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

const roomCode = ref(typeof route.query.room_code === 'string' ? route.query.room_code : '')
const nickname = ref('')

const isJoinDisabled = computed(() => {
  return roomCode.value.trim().length !== 8 || nickname.value.trim().length < 2
})

const handleJoin = async () => {
  const normalizedCode = roomCode.value.trim()
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

useHead({
  title: 'Вход в игру',
})
</script>

<template>
  <section class="home-entry">
    <Card class="w-full">
      <template #title>
        <div class="home-entry__title">Войти в игру</div>
      </template>

      <template #content>
        <div class="home-entry__fields">
          <FloatLabel variant="in">
            <InputMask
              id="room_code"
              v-model="roomCode"
              mask="99999999"
              slot-char=""
              class="w-full"
            />
            <label for="room_code">Код комнаты</label>
          </FloatLabel>

          <FloatLabel variant="in">
            <InputText id="nickname" v-model="nickname" maxlength="32" class="w-full" />
            <label for="nickname">Ваш ник</label>
          </FloatLabel>
        </div>
      </template>

      <template #footer>
        <div class="home-entry__actions">
          <Button label="Подключиться" :disabled="isJoinDisabled" @click="handleJoin" />
        </div>
      </template>
    </Card>
  </section>
</template>

<style scoped>
.home-entry {
  display: flex;
  width: min(100%, 24rem);
  flex-direction: column;
  gap: 0.75rem;
  margin-inline: auto;
}

.home-entry__title {
  margin-bottom: 0.5rem;
  text-align: center;
}

.home-entry__fields {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.home-entry__actions {
  display: flex;
  justify-content: center;
  padding-top: 0.75rem;
}
</style>
