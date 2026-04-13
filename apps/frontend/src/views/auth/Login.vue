<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import FloatLabel from 'primevue/floatlabel'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import { useToast } from 'primevue/usetoast'
import { useAuthStore } from '../../stores/auth'

const email = ref('')
const password = ref('')
const isSubmitting = ref(false)

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()
const toast = useToast()

type UiHttpError = Error & { status?: number }

const getLoginErrorMessage = (error: unknown): string => {
  const status = (error as UiHttpError)?.status

  if (status === 401) {
    return 'Неверный email или пароль.'
  }

  if (status === 422) {
    return 'Проверьте корректность email и пароля.'
  }

  if (error instanceof Error && error.message.trim().length > 0) {
    return error.message
  }

  return 'Не удалось выполнить вход. Попробуйте снова.'
}

const handleLogin = async (): Promise<void> => {
  if (isSubmitting.value) {
    return
  }

  const normalizedEmail = email.value.trim()
  const normalizedPassword = password.value.trim()

  if (!normalizedEmail || !normalizedPassword) {
    toast.add({
      group: 'global',
      severity: 'warn',
      summary: 'Не все поля заполнены',
      detail: 'Укажите email и пароль.',
      life: 3000,
    })
    return
  }

  isSubmitting.value = true

  try {
    await authStore.login({
      email: normalizedEmail,
      password: normalizedPassword,
    })

    const redirect = route.query.redirect
    const redirectTarget = typeof redirect === 'string' && redirect.length > 0 ? redirect : '/host'

    toast.add({
      group: 'global',
      severity: 'success',
      summary: 'Вход выполнен',
      detail: 'Добро пожаловать в панель управления!',
      life: 2500,
    })

    await router.replace(redirectTarget)
  } catch (error: unknown) {
    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Не удалось войти',
      detail: getLoginErrorMessage(error),
      life: 3500,
    })
  } finally {
    isSubmitting.value = false
  }
}
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
          <Button
            label="Login"
            :loading="isSubmitting"
            :disabled="isSubmitting"
            @click="handleLogin"
          />
        </div>
      </div>
    </template>

    <template #footer>
      <span class="text-xs">Doesn't have an account yet? </span>
      <router-link
        to="/register"
        class="text-xs font-bold text-[var(--app-color-primary)] transition-colors hover:text-[var(--app-color-primary-hover)]"
      >
        Register
      </router-link>
    </template>
  </Card>
</template>
