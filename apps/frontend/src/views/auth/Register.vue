<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import FloatLabel from 'primevue/floatlabel'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import { useToast } from 'primevue/usetoast'
import { useAuthStore } from '@/stores/auth'

const email = ref('')
const nickname = ref('')
const password = ref('')
const isSubmitting = ref(false)

const authStore = useAuthStore()
const router = useRouter()
const toast = useToast()

type UiHttpError = Error & { status?: number }

const getRegisterErrorMessage = (error: unknown): string => {
  const status = (error as UiHttpError)?.status

  if (status === 409) {
    return 'Пользователь с таким email уже существует.'
  }

  if (status === 422) {
    return 'Проверьте введённые данные регистрации.'
  }

  if (error instanceof Error && error.message.trim().length > 0) {
    return error.message
  }

  return 'Не удалось зарегистрироваться. Попробуйте снова.'
}

const handleRegister = async (): Promise<void> => {
  if (isSubmitting.value) {
    return
  }

  const normalizedEmail = email.value.trim()
  const normalizedNickname = nickname.value.trim()
  const normalizedPassword = password.value.trim()

  if (!normalizedEmail || !normalizedNickname || !normalizedPassword) {
    toast.add({
      group: 'global',
      severity: 'warn',
      summary: 'Не все поля заполнены',
      detail: 'Укажите email, nickname и пароль.',
      life: 3000,
    })
    return
  }

  isSubmitting.value = true

  try {
    await authStore.register({
      email: normalizedEmail,
      nickname: normalizedNickname,
      password: normalizedPassword,
    })

    await authStore.login({
      email: normalizedEmail,
      password: normalizedPassword,
    })

    toast.add({
      group: 'global',
      severity: 'success',
      summary: 'Аккаунт создан',
      detail: 'Вы успешно зарегистрированы и вошли в систему.',
      life: 3000,
    })

    await router.replace('/host')
  } catch (error: unknown) {
    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Ошибка регистрации',
      detail: getRegisterErrorMessage(error),
      life: 3500,
    })
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <Card class="auth-card">
    <template #title>Registration</template>

    <template #content>
      <div class="auth-form">
        <FloatLabel variant="in" class="w-full">
          <InputText id="register_email" v-model="email" autocomplete="off" class="w-full" />
          <label for="register_email">Email</label>
        </FloatLabel>

        <FloatLabel variant="in" class="w-full">
          <InputText id="register_nickname" v-model="nickname" autocomplete="off" class="w-full" />
          <label for="register_nickname">Nickname</label>
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

        <div class="auth-form__actions">
          <Button
            label="Register"
            :loading="isSubmitting"
            :disabled="isSubmitting"
            @click="handleRegister"
          />
        </div>
      </div>
    </template>

    <template #footer>
      <div class="auth-card__footer">
        <span class="auth-card__footer-text">Have an account already?</span>
        <router-link to="/login" class="auth-card__footer-link"> Login </router-link>
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
