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
const password = ref('')
const isSubmitting = ref(false)

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()
const toast = useToast()

const getErrorMessage = (error: unknown): string => {
  if (error instanceof ApiHttpError) {
    if (error.status === 401) {
      return 'Неверный email или пароль.'
    }

    if (error.status === 422) {
      return 'Проверьте корректность email и пароля.'
    }

    return error.message
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

    toast.add({
      group: 'global',
      severity: 'success',
      summary: 'Вход выполнен',
      detail: 'Добро пожаловать в панель управления!',
      life: 2500,
    })

    const redirect = route.query.redirect
    const target = typeof redirect === 'string' && redirect.length > 0 ? redirect : '/host'
    await router.replace(target)
  } catch (error: unknown) {
    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Не удалось войти',
      detail: getErrorMessage(error),
      life: 3500,
    })
  } finally {
    isSubmitting.value = false
  }
}

useHead({
  title: 'Вход',
})
</script>

<template>
  <Card class="mx-auto w-full max-w-(--app-card-narrow)">
    <template #title>Вход</template>

    <template #content>
      <div class="flex flex-col gap-3">
        <FloatLabel variant="in" class="w-full">
          <InputText id="login_email" v-model="email" autocomplete="off" class="w-full" />
          <label for="login_email">Почта</label>
        </FloatLabel>

        <FloatLabel variant="in" class="w-full">
          <Password
            v-model="password"
            :feedback="false"
            class="w-full"
            input-class="w-full"
            input-id="login_password"
          />
          <label for="login_password">Пароль</label>
        </FloatLabel>

        <div class="flex justify-center pt-2">
          <Button label="Войти" :loading="isSubmitting" :disabled="isSubmitting" @click="handleLogin" />
        </div>
      </div>
    </template>

    <template #footer>
      <div class="mt-2 flex flex-wrap items-center gap-1.5">
        <span class="text-xs text-(--app-color-text-muted)">Ещё нет аккаунта?</span>
        <NuxtLink
          to="/register"
          class="text-xs font-bold text-(--app-color-primary) no-underline transition-colors duration-200 hover:text-(--app-color-primary-hover)"
        >
          Зарегистрироваться
        </NuxtLink>
      </div>
    </template>
  </Card>
</template>
