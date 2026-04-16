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

</script>

<template>
  <section class="home-entry">
    <Card class="w-full">
      <template #title>
        <div class="home-entry__title">Войти в игру</div>
      </template>

      <template #content>
        <FloatLabel variant="in">
          <InputMask id="digits" v-model="gameCode" mask="999999" slotChar="" class="w-full" />
          <label for="digits">Код</label>
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
