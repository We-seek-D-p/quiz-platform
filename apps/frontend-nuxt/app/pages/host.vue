<script setup lang="ts">
import Card from 'primevue/card'
import Message from 'primevue/message'
import Skeleton from 'primevue/skeleton'
import { ApiHttpError } from '~/composables/api/useApiClient'
import { useManagementApi } from '~/composables/api/useManagementApi'
import { useAuthStore } from '~/stores/auth'
import type { QuizPublic } from '~/types/management'

definePageMeta({
  middleware: 'auth',
  layout: 'dashboard',
  title: 'Панель управления',
})

const authStore = useAuthStore()
const managementApi = useManagementApi()

const quizzes = ref<QuizPublic[]>([])
const isLoading = ref(false)
const errorMessage = ref('')

const lastUpdatedQuiz = computed(() => {
  if (quizzes.value.length === 0) {
    return null
  }

  return [...quizzes.value].sort((a, b) => {
    return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
  })[0]
})

const getErrorMessage = (error: unknown): string => {
  if (error instanceof ApiHttpError && error.message.trim().length > 0) {
    return error.message
  }

  if (error instanceof Error && error.message.trim().length > 0) {
    return error.message
  }

  return 'Не удалось загрузить данные панели.'
}

const loadDashboard = async (): Promise<void> => {
  if (!authStore.accessToken) {
    errorMessage.value = 'Сессия недоступна. Выполните вход снова.'
    quizzes.value = []
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    quizzes.value = await managementApi.getQuizzes()
  } catch (error: unknown) {
    errorMessage.value = getErrorMessage(error)
    quizzes.value = []
  } finally {
    isLoading.value = false
  }
}

onMounted(loadDashboard)

useHead({
  title: 'Панель управления',
})
</script>

<template>
  <section class="host-dashboard">
    <header class="host-dashboard__header">
      <p class="host-dashboard__greeting">Здравствуйте, {{ authStore.user?.nickname ?? 'хост' }}.</p>
      <p class="host-dashboard__subtitle">Обзор квизов и последних изменений.</p>
    </header>

    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>

    <div class="host-dashboard__grid" v-if="isLoading">
      <Card>
        <template #content>
          <Skeleton width="100%" height="5rem" />
        </template>
      </Card>
      <Card>
        <template #content>
          <Skeleton width="100%" height="5rem" />
        </template>
      </Card>
    </div>

    <div class="host-dashboard__grid" v-else>
      <Card>
        <template #title>Всего квизов</template>
        <template #content>
          <p class="host-dashboard__metric">{{ quizzes.length }}</p>
        </template>
      </Card>

      <Card>
        <template #title>Последнее обновление</template>
        <template #content>
          <p class="host-dashboard__metric host-dashboard__metric--small">
            {{ lastUpdatedQuiz ? lastUpdatedQuiz.title : 'Квизы еще не созданы' }}
          </p>
        </template>
      </Card>
    </div>
  </section>
</template>

<style scoped>
.host-dashboard {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.host-dashboard__header {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.host-dashboard__greeting {
  margin: 0;
  font-size: 2rem;
  font-weight: 700;
  line-height: 1.15;
}

.host-dashboard__subtitle {
  margin: 0;
  color: var(--app-color-text-muted);
}

.host-dashboard__grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(1, minmax(0, 1fr));
}

.host-dashboard__metric {
  margin: 0;
  font-size: 2.25rem;
  font-weight: 700;
}

.host-dashboard__metric--small {
  font-size: 1.125rem;
}

@media (min-width: 900px) {
  .host-dashboard__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .host-dashboard__greeting {
    font-size: 2.25rem;
  }
}
</style>
