<script setup lang="ts">
import {ref} from 'vue'
import Button from 'primevue/button'
import Card from 'primevue/card'
import FloatLabel from 'primevue/floatlabel';
import InputNumber from 'primevue/inputnumber'
import { useRouter } from 'vue-router'
import {getStoredThemeMode, toggleThemeMode, type ThemeMode} from '../../theme.ts'

const count = ref(0)
const themeMode = ref<ThemeMode>(getStoredThemeMode())
const router = useRouter()

const increment = () => {
  count.value += 1
}

const goToDashboard = () => {
  router.push('/host')
}

const switchTheme = () => {
  themeMode.value = toggleThemeMode()
}
</script>

<template>
  <main class="app-root">
    <Card class="demo-card p-anchored-overlay-enter-active">
      <template #title >
        <div class="text-center mb-2">Join a game</div>
      </template>
      <template #content>
        <FloatLabel variant="in">
          <InputNumber :useGrouping="false" fluid />
          <label>Code</label>

        </FloatLabel>
      </template>
      <template #footer>
        <div class="flex justify-center w-full mt-3">
          <Button label="Join"  @click="increment"/>
        </div>
      </template>
    </Card>

    <Button
        class="theme-button"
        :label="`Theme: ${themeMode}`"
        icon="pi pi-palette"
        @click="switchTheme"
    />
    <Button
         class="theme-button"
         :label="`Host dashboard`"
         @click="goToDashboard"
    />
    <Button
        class="theme-button"
        :label="`Register page`"
        @click="router.push('/register')"
    />
  </main>
</template>

<style scoped>
.app-root {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 1rem;
}

.demo-card {
  width: min(28rem, 100%);
}

.count-text {
  margin: 0;
}

.theme-button {
  width: min(28rem, 100%);
}
</style>
