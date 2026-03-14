<script setup lang="ts">
import {ref} from 'vue'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Password from 'primevue/password'
import InputText from 'primevue/inputtext'
import {getStoredThemeMode, toggleThemeMode, type ThemeMode} from '../theme'

const count = ref(0)
const themeMode = ref<ThemeMode>(getStoredThemeMode())

const increment = () => {
  count.value += 1
}

const switchTheme = () => {
  themeMode.value = toggleThemeMode()
}

const SendRegister = (): void => {}

</script>

<template>
  <main class="app-root">
    <Card class="demo-card">
      <template #title>Register</template>
      <template #content>
        <p class="count-text">Email</p>
        <InputText type="text" v-model="value" />
        <p class="count-text">Password</p>
        <Password v-model="value" :feedback="false" />
        <p class="count-text">Confirm password</p>
        <Password v-model="value" :feedback="false" />


        <span>Already registered? </span>
        <Link label="login" @click="SendRegister"/>
      </template>
      <template #footer>
        <router-link to="/login">Login</router-link>
      </template>
    </Card>

    <Button
        class="theme-button"
        :label="`Theme: ${themeMode}`"
        icon="pi pi-palette"
        @click="switchTheme"
    />
  </main>
</template>

<style scoped>
.app-root {
  min-height: 100vh;
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
