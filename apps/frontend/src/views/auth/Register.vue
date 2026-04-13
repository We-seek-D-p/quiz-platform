<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import FloatLabel from 'primevue/floatlabel'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import { useToast } from 'primevue/usetoast'
import { useAuthStore } from '@/stores/auth.ts'

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
  <Card>
    <template #title>Registration</template>

    <template #content>
      <div class="space-y-2">
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

        <div class="flex justify-center pt-2">
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
