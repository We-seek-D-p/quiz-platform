<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import FloatLabel from 'primevue/floatlabel'
import InputMask from 'primevue/inputmask'

const router = useRouter()
const gameCode = ref<number | null>(null)

const handleJoin = () => {}
const goToDashboard = () => {
  router.push('/host')
}
const goToRegister = () => {
  router.push('/register')
}
</script>

<template>
  <section class="flex flex-col gap-3">
    <Card class="w-full">
      <template #title>
        <div class="mb-2 text-center">Join a game</div>
      </template>

      <template #content>
        <FloatLabel variant="in">
          <InputMask id="digits" v-model="value" mask="999999" slotChar="" />
          <label for="digits">Code</label>
        </FloatLabel>
      </template>

      <template #footer>
        <div class="flex justify-center pt-3">
          <Button label="Join" @click="handleJoin" />
        </div>
      </template>
    </Card>

    <Button class="w-full" label="Host dashboard" @click="goToDashboard" />
    <Button class="w-full" label="Register page" @click="goToRegister" />
  </section>
</template>
