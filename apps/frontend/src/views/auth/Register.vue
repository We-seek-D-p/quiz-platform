<script setup lang="ts">
import { ref } from 'vue'
import Button from 'primevue/button'
import Card from 'primevue/card'
import FloatLabel from 'primevue/floatlabel'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'

const email = ref('')
const username = ref('')
const password = ref('')
const confirmPassword = ref('')

const handleRegister = (): void => {}
</script>

<template>
  <Card>
    <template #title>Registration</template>

    <template #content>
      <div class="space-y-2">
        <FloatLabel variant="in" class="w-full">
          <InputText id="register_email" v-model="email" autocomplete="off" class="w-full" />
          <label for="register_email">Email</label>
        </FloatLabel>

        <FloatLabel variant="in" class="w-full">
          <InputText id="register_username" v-model="username" autocomplete="off" class="w-full" />
          <label for="register_username">Username</label>
        </FloatLabel>

        <FloatLabel variant="in" class="w-full">
          <Password
            v-model="password"
            :feedback="false"
            class="w-full"
            input-class="w-full"
            input-id="register_password"
          />
          <label for="register_password">Password</label>
        </FloatLabel>

        <FloatLabel variant="in" class="w-full">
          <Password
            v-model="confirmPassword"
            :feedback="false"
            class="w-full"
            input-class="w-full"
            input-id="register_confirm_password"
          />
          <label for="register_confirm_password">Confirm password</label>
        </FloatLabel>

        <div class="flex justify-center pt-2">
          <Button label="Register" @click="handleRegister" />
        </div>
      </div>
    </template>

    <template #footer>
      <span class="text-xs">Have an account already? </span>
      <router-link
        to="/login"
        class="text-xs font-bold text-[var(--app-color-primary)] transition-colors hover:text-[var(--app-color-primary-hover)]"
      >
        Login
      </router-link>
    </template>
  </Card>
</template>
