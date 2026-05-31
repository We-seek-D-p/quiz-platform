<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Dialog from 'primevue/dialog'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import { ApiHttpError } from '~/composables/api/useApiClient'
import { useManagementApi } from '~/composables/api/useManagementApi'
import { useAuthStore } from '~/stores/auth'
import type { QuizPublic } from '~/types/management'

type QuizTableRow = QuizPublic & {
  questionCount: number
}

definePageMeta({
  middleware: 'auth',
  layout: 'dashboard',
  title: 'Мои квизы',
})

const authStore = useAuthStore()
const router = useRouter()
const toast = useToast()
const managementApi = useManagementApi()

const quizzes = ref<QuizTableRow[]>([])
const isLoading = ref(false)
const errorMessage = ref('')
const deletingQuizId = ref<string | null>(null)
const creatingSessionQuizId = ref<string | null>(null)
const quizPendingDeletion = ref<QuizTableRow | null>(null)

const hasAccessContext = computed(() => {
  return Boolean(authStore.accessToken)
})

const isEmpty = computed(() => {
  return !isLoading.value && !errorMessage.value && quizzes.value.length === 0
})

const formatDate = (value: string): string => {
  return new Date(value).toLocaleDateString()
}

const ERROR_CODE_MESSAGES: Record<string, string> = {
  unauthorized: 'Сессия истекла. Войдите снова',
  forbidden: 'Недостаточно прав для этого действия',
  quiz_not_ready: 'Добавьте вопросы перед запуском сессии',
  session_provider_unavailable: 'Session Service временно недоступен',
  session_provider_error: 'Не удалось инициализировать игровую сессию',
}

const handleUnauthorizedError = async (error: unknown) => {
  if (!(error instanceof ApiHttpError) || error.status !== 401) {
    return false
  }

  await router.push('/login')
  return true
}

const getErrorMessage = (error: unknown, fallback: string): string => {
  if (error instanceof ApiHttpError) {
    const knownMessage = error.code ? ERROR_CODE_MESSAGES[error.code] : undefined
    if (knownMessage) {
      return knownMessage
    }

    if (error.message.trim().length > 0) {
      return error.message
    }
  }

  if (error instanceof Error && error.message.trim().length > 0) {
    return error.message
  }

  return fallback
}

const fetchQuizzes = async (): Promise<void> => {
  if (!authStore.accessToken) {
    errorMessage.value = 'Сессия недоступна. Выполните вход снова'
    quizzes.value = []
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    const quizList = await managementApi.getQuizzes()

    const questionCounts = await Promise.all(
      quizList.map(async (quiz) => {
        try {
          const questions = await managementApi.getQuizQuestions(quiz.id)
          return questions.length
        } catch {
          return 0
        }
      }),
    )

    quizzes.value = quizList.map((quiz, index) => {
      return {
        ...quiz,
        questionCount: questionCounts[index] ?? 0,
      }
    })
  } catch (error: unknown) {
    if (await handleUnauthorizedError(error)) {
      errorMessage.value = 'Сессия истекла. Войдите снова'
      quizzes.value = []
      return
    }

    errorMessage.value = getErrorMessage(error, 'Не удалось загрузить список квизов')
    quizzes.value = []
  } finally {
    isLoading.value = false
  }
}

const openCreateQuiz = async (): Promise<void> => {
  await router.push('/quizzes/editor')
}

const editQuiz = async (quizId: string): Promise<void> => {
  await router.push(`/quizzes/editor?quiz=${quizId}`)
}

const launchQuizSession = async (quiz: QuizTableRow): Promise<void> => {
  if (!authStore.accessToken || creatingSessionQuizId.value) {
    return
  }

  creatingSessionQuizId.value = quiz.id

  try {
    const session = await managementApi.createSession({
      quiz_id: quiz.id,
    })

    if (!session.id) {
      throw new Error('Session id отсутствует в ответе сервера')
    }

    if (!session.room_code) {
      toast.add({
        group: 'global',
        severity: 'warn',
        summary: 'Сессия не готова',
        detail: 'Сервис не вернул room code. Попробуйте создать сессию снова.',
        life: 3500,
      })
      return
    }

    const query: Record<string, string> = {
      session_id: session.id,
      room_code: session.room_code,
    }

    await router.push({
      path: '/host',
      query,
    })
  } catch (error: unknown) {
    if (await handleUnauthorizedError(error)) {
      return
    }

    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Не удалось создать сессию',
      detail: getErrorMessage(error, 'Попробуйте снова'),
      life: 3500,
    })
  } finally {
    creatingSessionQuizId.value = null
  }
}

const askDeleteQuiz = (quiz: QuizTableRow): void => {
  quizPendingDeletion.value = quiz
}

const closeDeleteDialog = (): void => {
  if (deletingQuizId.value) {
    return
  }

  quizPendingDeletion.value = null
}

const confirmDeleteQuiz = async (): Promise<void> => {
  const target = quizPendingDeletion.value
  if (!target || !authStore.accessToken) {
    return
  }

  deletingQuizId.value = target.id

  try {
    await managementApi.deleteQuiz(target.id)
    quizzes.value = quizzes.value.filter((quiz) => quiz.id !== target.id)
    quizPendingDeletion.value = null

    toast.add({
      group: 'global',
      severity: 'success',
      summary: 'Квиз удален',
      detail: `"${target.title}" удален из списка`,
      life: 2500,
    })
  } catch (error: unknown) {
    if (await handleUnauthorizedError(error)) {
      return
    }

    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Не удалось удалить квиз',
      detail: getErrorMessage(error, 'Попробуйте снова'),
      life: 3500,
    })
  } finally {
    deletingQuizId.value = null
  }
}

onMounted(fetchQuizzes)

useHead({
  title: 'Мои квизы',
})
</script>

<template>
  <section class="flex flex-col gap-(--app-page-gap)">
    <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
      <div>
        <p class="m-0 text-(--app-color-text-muted)">Управляйте контентом и переходите в редактор одним кликом.</p>
      </div>

      <Button label="Новый квиз" icon="pi pi-plus" :disabled="!hasAccessContext" @click="openCreateQuiz" />
    </div>

    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>

    <Card>
      <template #content>
        <DataTable :value="quizzes" :loading="isLoading" class="quizzes-table">
          <Column field="title" header="Название" />

          <Column header="Вопросов">
            <template #body="slotProps">
              <Tag :value="slotProps.data.questionCount" severity="secondary" />
            </template>
          </Column>

          <Column header="Создан">
            <template #body="slotProps">
              {{ formatDate(slotProps.data.created_at) }}
            </template>
          </Column>

          <Column header-style="text-align: right" body-style="text-align: right">
            <template #body="slotProps">
              <div class="inline-flex gap-1">
                <Button
                  icon="pi pi-play"
                  text
                  severity="success"
                  aria-label="Запустить игру"
                  :loading="creatingSessionQuizId === slotProps.data.id"
                  :disabled="Boolean(creatingSessionQuizId)"
                  @click="launchQuizSession(slotProps.data)"
                />
                <Button
                  icon="pi pi-pencil"
                  text
                  severity="secondary"
                  aria-label="Редактировать квиз"
                  @click="editQuiz(slotProps.data.id)"
                />
                <Button
                  icon="pi pi-trash"
                  text
                  severity="danger"
                  aria-label="Удалить квиз"
                  @click="askDeleteQuiz(slotProps.data)"
                />
              </div>
            </template>
          </Column>

          <template #empty>
            <div v-if="isEmpty" class="p-5 text-center text-(--app-color-text-muted)">
              У вас пока нет квизов. Создайте первый квиз.
            </div>
          </template>
        </DataTable>
      </template>
    </Card>

    <Dialog
      :visible="Boolean(quizPendingDeletion)"
      modal
      header="Удалить квиз"
      :draggable="false"
      :closable="!deletingQuizId"
      @update:visible="closeDeleteDialog"
    >
      <p class="m-0">
        Квиз
        <strong>{{ quizPendingDeletion?.title }}</strong>
        будет удален без возможности восстановления.
      </p>

      <template #footer>
        <div class="flex justify-end gap-2">
          <Button label="Отмена" text :disabled="Boolean(deletingQuizId)" @click="closeDeleteDialog" />
          <Button
            label="Удалить"
            severity="danger"
            :loading="Boolean(deletingQuizId)"
            :disabled="Boolean(deletingQuizId)"
            @click="confirmDeleteQuiz"
          />
        </div>
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.quizzes-table :deep(.p-datatable-table) {
  min-width: 42rem;
}
</style>
