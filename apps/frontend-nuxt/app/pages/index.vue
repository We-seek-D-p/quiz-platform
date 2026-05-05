<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import FloatLabel from 'primevue/floatlabel'
import InputMask from 'primevue/inputmask'
import { useToast } from 'primevue/usetoast'

const roomCode = ref('')
const toast = useToast()

const handleJoin = () => {
  const normalizedCode = roomCode.value.trim()

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

  toast.add({
    group: 'global',
    severity: 'info',
    summary: 'Недоступно',
    detail: 'Подключение по коду комнаты пока недоступно',
    life: 3000,
  })
}

useHead({
  title: 'Главная',
})
</script>

<template>
  <section class="home-entry">
    <Card class="w-full">
      <template #title>
        <div class="home-entry__title">Войти в игру</div>
      </template>

      <template #content>
        <FloatLabel variant="in">
          <InputMask id="room_code" v-model="roomCode" mask="99999999" slot-char="" class="w-full" />
          <label for="room_code">Код комнаты</label>
        </FloatLabel>
      </template>

      <template #footer>
        <div class="home-entry__actions">
          <Button label="Подключиться" @click="handleJoin" />
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

.home-entry__actions {
  display: flex;
  justify-content: center;
  padding-top: 0.75rem;
}
</style>
