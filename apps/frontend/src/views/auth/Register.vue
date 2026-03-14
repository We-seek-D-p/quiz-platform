<script setup lang="ts">
import {ref} from 'vue'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Password from 'primevue/password'
import InputText from 'primevue/inputtext'
import FloatLabel from 'primevue/floatlabel'
import { useRouter } from 'vue-router'
import {getStoredThemeMode, toggleThemeMode, type ThemeMode} from '../../theme.ts'

const count = ref(0)
const themeMode = ref<ThemeMode>(getStoredThemeMode())
const router = useRouter()

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
        <FloatLabel variant="in">
          <InputText id="email" v-model="value" autocomplete="off" />
          <label for="email">Email</label>
        </FloatLabel>
        <FloatLabel variant="in">
          <InputText id="username" v-model="value" autocomplete="off" />
          <label for="username">Username</label>
        </FloatLabel>
        <FloatLabel variant="in">
        <Password v-model="value" :feedback="false" />
        <label>Password</label>
        </FloatLabel>
        <FloatLabel variant="in">
        <Password v-model="value" :feedback="false" />
        <label>Confirm password</label>
        </FloatLabel>
        <span>Doesn't have an account yet? </span>
        <Link label="login" @click="SendLogin"/>
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
    <Button
        class="theme-button"
        :label="`Login page`"
        @click="router.push('/login')"
    />
    <Button
        class="theme-button"
        :label="`Main page`"
        @click="router.push('/')"
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
