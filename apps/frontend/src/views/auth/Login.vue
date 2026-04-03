<script setup lang="ts">
import { ref } from 'vue'
import Button from 'primevue/button'
import Card from 'primevue/card'
import FloatLabel from 'primevue/floatlabel'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'

const email = ref('')
const password = ref('')

const handleLogin = (): void => {}
</script>

<template>
  <Card>
    <template #title>Login</template>

    <template #content>
      <div class="space-y-3">
        <FloatLabel variant="in" class="w-full">
          <InputText id="login_email" v-model="email" autocomplete="off" class="w-full" />
          <label for="login_email">Email</label>
        </FloatLabel>

        <FloatLabel variant="in" class="w-full">
          <Password
            v-model="password"
            :feedback="false"
            class="w-full"
            input-class="w-full"
            input-id="login_password"
          />
          <label for="login_password">Password</label>
        </FloatLabel>

        <div class="flex justify-center pt-2">
          <Button label="Login" @click="handleLogin" />
        </div>
      </div>
    </template>

    <template #footer>
      <span class="text-xs ">Doesn't have an account yet? </span>
      <router-link
        to="/register"
        class="text-xs font-bold text-[var(--app-color-primary)] transition-colors hover:text-[var(--app-color-primary-hover)]"
      >
        Register
      </router-link>
    </template>
  </Card>
</template>
