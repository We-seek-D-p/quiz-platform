<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import FloatLabel from 'primevue/floatlabel'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import { useToast } from 'primevue/usetoast'
import { ApiHttpError } from '~/composables/api/useApiClient'
import { useAuthStore } from '~/stores/auth'

definePageMeta({
  middleware: 'guest',
})

const email = ref('')
const nickname = ref('')
const password = ref('')
const isSubmitting = ref(false)

const authStore = useAuthStore()
const router = useRouter()
const toast = useToast()

const getErrorMessage = (error: unknown): string => {
  if (error instanceof ApiHttpError) {
    if (error.status === 409) {
      return 'Пользователь с таким email уже существует.'
    }

    if (error.status === 422) {
      return 'Проверьте введённые данные регистрации.'
    }

    return error.message
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
      detail: getErrorMessage(error),
      life: 3500,
    })
  } finally {
    isSubmitting.value = false
  }
}

useHead({
  title: 'Регистрация',
})
</script>

<template>
  <Card class="mx-auto w-full max-w-(--app-card-narrow)">
    <template #title>Создать аккаунт</template>

    <template #content>
      <div class="flex flex-col gap-3">
        <FloatLabel variant="in" class="w-full">
          <InputText id="register_email" v-model="email" autocomplete="off" class="w-full" />
          <label for="register_email">Почта</label>
        </FloatLabel>

        <FloatLabel variant="in" class="w-full">
          <InputText id="register_nickname" v-model="nickname" autocomplete="off" class="w-full" />
          <label for="register_nickname">Логин</label>
        </FloatLabel>

        <FloatLabel variant="in" class="w-full">
          <Password
            v-model="password"
            :feedback="false"
            class="w-full"
            input-class="w-full"
            input-id="register_password"
          />
          <label for="register_password">Пароль</label>
        </FloatLabel>

        <div class="flex justify-center pt-2">
          <Button label="Зарегистрироваться" :loading="isSubmitting" :disabled="isSubmitting" @click="handleRegister" />
        </div>
      </div>
    </template>

    <template #footer>
      <div class="mt-2 flex flex-wrap items-center gap-1.5">
        <span class="text-xs text-(--app-color-text-muted)">Уже есть аккаунт?</span>
        <NuxtLink
          to="/login"
          class="text-xs font-bold text-(--app-color-primary) no-underline transition-colors duration-200 hover:text-(--app-color-primary-hover)"
        >
          Войти
        </NuxtLink>
      </div>
    </template>
  </Card>
</template>
