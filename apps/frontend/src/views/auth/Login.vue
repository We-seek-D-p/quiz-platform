<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import FloatLabel from 'primevue/floatlabel'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import { useToast } from 'primevue/usetoast'
import { useAuthStore } from '@/stores/auth'

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
  <Card class="auth-card">
    <template #title>Login</template>

    <template #content>
      <div class="auth-form">
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

        <div class="auth-form__actions">
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
      <div class="auth-card__footer">
        <span class="auth-card__footer-text">Doesn't have an account yet?</span>
        <router-link to="/register" class="auth-card__footer-link"> Register </router-link>
      </div>
    </template>
  </Card>
</template>

<style scoped>
.auth-card {
  width: min(100%, 24rem);
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.auth-form__actions {
  display: flex;
  justify-content: center;
  padding-top: 0.5rem;
}

.auth-card__footer {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  flex-wrap: wrap;
}

.auth-card__footer-text {
  font-size: 0.75rem;
}

.auth-card__footer-link {
  color: var(--app-color-primary);
  font-size: 0.75rem;
  font-weight: 700;
  text-decoration: none;
  transition: color var(--app-transition-fast);
}

.auth-card__footer-link:hover {
  color: var(--app-color-primary-hover);
}
</style>
