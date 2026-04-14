<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import FloatLabel from 'primevue/floatlabel'
import InputMask from 'primevue/inputmask'

const router = useRouter()
const gameCode = ref('')

const handleJoin = () => {
  // TODO: implement join flow
}

const goToDashboard = () => {
  router.push('/host')
}

const goToRegister = () => {
  router.push('/register')
}
</script>

<template>
  <section class="home-entry">
    <Card class="w-full">
      <template #title>
        <div class="home-entry__title">Join a game</div>
      </template>

      <template #content>
        <FloatLabel variant="in">
          <InputMask id="digits" v-model="gameCode" mask="999999" />
          <label for="digits">Code</label>
        </FloatLabel>
      </template>

      <template #footer>
        <div class="home-entry__actions">
          <Button label="Join" @click="handleJoin" />
        </div>
      </template>
    </Card>

    <Button class="w-full" label="Host dashboard" @click="goToDashboard" />
    <Button class="w-full" label="Register page" @click="goToRegister" />
  </section>
</template>

<style scoped>
.home-entry {
  display: flex;
  width: min(100%, 24rem);
  flex-direction: column;
  gap: 0.75rem;
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
