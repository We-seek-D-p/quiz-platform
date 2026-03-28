<script setup lang="ts">
import {ref} from 'vue'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Password from 'primevue/password'
import InputText from 'primevue/inputtext'
import FloatLabel from 'primevue/floatlabel'
import {getStoredThemeMode, toggleThemeMode, type ThemeMode} from '../../theme.ts'

const themeMode = ref<ThemeMode>(getStoredThemeMode())

const SendRegister = (): void => {}

</script>

<template>
  <main class="app-root">
    <Card class="demo-card p-anchored-overlay-enter-active">
      <template #title>Registration</template>
      <template #content>
        <FloatLabel variant="in" class="m-3">
          <InputText id="email" v-model="value" autocomplete="off" />
          <label for="email">Email</label>
        </FloatLabel>

        <FloatLabel variant="in" class="m-3">
          <InputText id="username" v-model="value" autocomplete="off" />
          <label for="username">Username</label>
        </FloatLabel>

        <FloatLabel variant="in" class="m-3">
        <Password v-model="value" :feedback="false" />
        <label>Password</label>
        </FloatLabel>

        <FloatLabel variant="in" class="m-3">
        <Password v-model="value" :feedback="false" />
        <label>Confirm password</label>
        </FloatLabel>

        <div class="flex justify-center w-full mt-4 mb-1">
          <Button label="Register"  @click="SendRegister"/>
        </div>

      </template>
      <template #footer>
        <span class="text-xs">Have an account already? </span>
        <router-link to="/login"><span class="text-xs font-bold text-primary cursor-pointer select-none">Login</span></router-link>
      </template>
    </Card>
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
