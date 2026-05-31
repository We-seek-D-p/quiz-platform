<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import Checkbox from 'primevue/checkbox'
import Dialog from 'primevue/dialog'
import Divider from 'primevue/divider'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import RadioButton from 'primevue/radiobutton'
import SelectButton from 'primevue/selectbutton'
import Textarea from 'primevue/textarea'
import { useToast } from 'primevue/usetoast'
import { VueDraggable } from 'vue-draggable-plus'
import { ApiHttpError } from '~/composables/api/useApiClient'
import { useManagementApi } from '~/composables/api/useManagementApi'
import { useAuthStore } from '~/stores/auth'
import type { QuestionCreate, QuestionPublic, QuestionUpdate, QuizUpdate } from '~/types/management'
import type { OptionDraft, QuestionDraft, QuizDraft } from '~/types/quiz-editor'

definePageMeta({
  middleware: 'auth',
  layout: 'dashboard',
  title: 'Редактор квиза',
})

const route = useRoute()
const router = useRouter()
const toast = useToast()
const authStore = useAuthStore()
const managementApi = useManagementApi()

const quiz = ref<QuizDraft>({
  title: '',
  description: '',
  questions: [],
})

const selectionOptions = [
  { label: 'Один ответ', value: 'single' },
  { label: 'Несколько', value: 'multiple' },
]

const isLoading = ref(false)
const isSaving = ref(false)
const showCancelDialog = ref(false)
const initialSnapshot = ref('')
const initialQuestionIds = ref<Set<string>>(new Set())

const generateLocalId = (): string => {
  return Math.random().toString(36).slice(2, 10)
}

const getErrorMessage = (error: unknown, fallback: string): string => {
  if (error instanceof ApiHttpError && error.message.trim().length > 0) {
    return error.message
  }

  if (error instanceof Error && error.message.trim().length > 0) {
    return error.message
  }

  return fallback
}

const getEditorQuizId = (): string | null => {
  const quizId = route.query.quiz
  if (typeof quizId === 'string' && quizId.length > 0 && quizId !== 'new') {
    return quizId
  }
  return null
}

const toDraftQuestion = (question: QuestionPublic): QuestionDraft => {
  return {
    localId: generateLocalId(),
    id: question.id,
    text: question.text,
    selection_type: question.selection_type,
    time_limit_seconds: question.time_limit_seconds,
    order_index: question.order_index,
    options: question.options
      .slice()
      .sort((left, right) => left.order_index - right.order_index)
      .map((option) => ({
        localId: generateLocalId(),
        id: option.id,
        text: option.text,
        is_correct: option.is_correct,
        order_index: option.order_index,
      })),
  }
}

const serializeDraft = (draft: QuizDraft): string => {
  return JSON.stringify({
    title: draft.title.trim(),
    description: draft.description.trim(),
    questions: draft.questions.map((question) => ({
      id: question.id ?? null,
      text: question.text.trim(),
      selection_type: question.selection_type,
      time_limit_seconds: question.time_limit_seconds,
      order_index: question.order_index,
      options: question.options.map((option) => ({
        id: option.id ?? null,
        text: option.text.trim(),
        is_correct: option.is_correct,
        order_index: option.order_index,
      })),
    })),
  })
}

const addQuestion = (): void => {
  const questionDraft: QuestionDraft = {
    localId: generateLocalId(),
    text: '',
    selection_type: 'single',
    time_limit_seconds: 20,
    order_index: quiz.value.questions.length,
    options: Array.from({ length: 4 }, (_, index) => ({
      localId: generateLocalId(),
      text: '',
      is_correct: index === 0,
      order_index: index,
    })),
  }

  quiz.value.questions.push(questionDraft)
}

const reorderQuestions = (): void => {
  quiz.value.questions = quiz.value.questions.map((question, index) => ({
    ...question,
    order_index: index,
  }))
}

const reorderQuestionOptions = (question: QuestionDraft): void => {
  question.options = question.options.map((option, index) => ({
    ...option,
    order_index: index,
  }))
}

const handleQuestionDragEnd = (): void => {
  reorderQuestions()
}

const handleOptionDragEnd = (question: QuestionDraft): void => {
  reorderQuestionOptions(question)
}

const removeQuestion = (localId: string): void => {
  quiz.value.questions = quiz.value.questions.filter((question) => question.localId !== localId)
  reorderQuestions()
}

const toggleCorrect = (question: QuestionDraft, option: OptionDraft): void => {
  if (question.selection_type === 'single') {
    question.options.forEach((candidate) => {
      candidate.is_correct = candidate.localId === option.localId
    })
    return
  }

  option.is_correct = !option.is_correct
}

const switchQuestionType = (question: QuestionDraft): void => {
  if (question.selection_type === 'single') {
    const firstCorrectIndex = question.options.findIndex((option) => option.is_correct)
    question.options.forEach((option, index) => {
      option.is_correct = index === (firstCorrectIndex >= 0 ? firstCorrectIndex : 0)
    })
  }
}

const normalizeQuestion = (question: QuestionDraft, orderIndex: number): QuestionDraft => {
  const normalizedOptions = question.options.map((option, optionIndex) => ({
    ...option,
    text: option.text.trim(),
    order_index: optionIndex,
  }))

  if (question.selection_type === 'single') {
    const firstCorrect = normalizedOptions.findIndex((option) => option.is_correct)
    normalizedOptions.forEach((option, index) => {
      option.is_correct = index === (firstCorrect >= 0 ? firstCorrect : 0)
    })
  }

  return {
    ...question,
    text: question.text.trim(),
    order_index: orderIndex,
    options: normalizedOptions,
  }
}

const validateQuizDraft = (): string | null => {
  if (quiz.value.title.trim().length === 0) {
    return 'Укажите название квиза'
  }

  if (quiz.value.questions.length === 0) {
    return 'Добавьте хотя бы один вопрос'
  }

  for (const [questionIndex, question] of quiz.value.questions.entries()) {
    if (question.text.trim().length === 0) {
      return `Заполните текст вопроса #${questionIndex + 1}`
    }

    if (!question.options.some((option) => option.is_correct)) {
      return `Выберите правильный ответ для вопроса #${questionIndex + 1}`
    }

    for (const [optionIndex, option] of question.options.entries()) {
      if (option.text.trim().length === 0) {
        return `Заполните вариант #${optionIndex + 1} в вопросе #${questionIndex + 1}`
      }
    }
  }

  return null
}

const buildQuestionCreatePayload = (question: QuestionDraft): QuestionCreate => {
  return {
    text: question.text,
    selection_type: question.selection_type,
    time_limit_seconds: question.time_limit_seconds,
    order_index: question.order_index,
    options: question.options.map((option) => ({
      text: option.text,
      order_index: option.order_index,
      is_correct: option.is_correct,
    })),
  }
}

const buildQuestionUpdatePayload = (question: QuestionDraft): QuestionUpdate => {
  return {
    text: question.text,
    selection_type: question.selection_type,
    time_limit_seconds: question.time_limit_seconds,
    order_index: question.order_index,
    options: question.options.map((option) => ({
      id: option.id,
      text: option.text,
      order_index: option.order_index,
      is_correct: option.is_correct,
    })),
  }
}

const hasUnsavedChanges = computed(() => {
  return serializeDraft(quiz.value) !== initialSnapshot.value
})

const loadQuiz = async (): Promise<void> => {
  const quizId = getEditorQuizId()
  if (!authStore.accessToken) {
    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Сессия недоступна',
      detail: 'Выполните вход снова.',
      life: 3000,
    })
    await router.replace('/login')
    return
  }

  if (!quizId) {
    addQuestion()
    initialSnapshot.value = serializeDraft(quiz.value)
    return
  }

  isLoading.value = true

  try {
    const [quizPayload, questions] = await Promise.all([
      managementApi.getQuiz(quizId),
      managementApi.getQuizQuestions(quizId),
    ] as const)

    quiz.value = {
      id: quizPayload.id,
      title: quizPayload.title,
      description: quizPayload.description,
      questions: questions
        .slice()
        .sort((left, right) => left.order_index - right.order_index)
        .map((question) => toDraftQuestion(question)),
    }

    initialQuestionIds.value = new Set(
      quiz.value.questions
        .map((question) => question.id)
        .filter((questionId): questionId is string => Boolean(questionId)),
    )

    if (quiz.value.questions.length === 0) {
      addQuestion()
    }

    initialSnapshot.value = serializeDraft(quiz.value)
  } catch (error: unknown) {
    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Не удалось загрузить квиз',
      detail: getErrorMessage(error, 'Попробуйте открыть страницу снова.'),
      life: 3500,
    })
    await router.replace('/quizzes')
  } finally {
    isLoading.value = false
  }
}

const handleCancel = async (): Promise<void> => {
  if (isSaving.value) {
    return
  }

  if (!hasUnsavedChanges.value) {
    await router.push('/quizzes')
    return
  }

  showCancelDialog.value = true
}

const confirmCancel = async (): Promise<void> => {
  showCancelDialog.value = false
  await router.push('/quizzes')
}

const handleSave = async (): Promise<void> => {
  const validationError = validateQuizDraft()
  if (validationError) {
    toast.add({
      group: 'global',
      severity: 'warn',
      summary: 'Проверьте форму',
      detail: validationError,
      life: 3200,
    })
    return
  }

  if (!authStore.accessToken) {
    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Сессия недоступна',
      detail: 'Выполните вход снова.',
      life: 3000,
    })
    return
  }

  isSaving.value = true
  const normalizedQuestions = quiz.value.questions.map((question, index) => normalizeQuestion(question, index))

  try {
    const quizPayload: QuizUpdate = {
      title: quiz.value.title.trim(),
      description: quiz.value.description.trim(),
    }

    let targetQuizId = quiz.value.id

    if (targetQuizId) {
      await managementApi.updateQuiz(targetQuizId, quizPayload)
    } else {
      const createdQuiz = await managementApi.createQuiz({
        title: quizPayload.title ?? '',
        description: quizPayload.description ?? '',
      })

      targetQuizId = createdQuiz.id
      quiz.value.id = createdQuiz.id
    }

    if (!targetQuizId) {
      throw new Error('Не удалось определить идентификатор квиза.')
    }

    const currentExistingIds = new Set(
      normalizedQuestions
        .map((question) => question.id)
        .filter((questionId): questionId is string => Boolean(questionId)),
    )
    const deletedQuestionIds = [...initialQuestionIds.value].filter((questionId) => !currentExistingIds.has(questionId))

    for (const deletedQuestionId of deletedQuestionIds) {
      await managementApi.deleteQuestion(targetQuizId, deletedQuestionId)
    }

    for (const question of normalizedQuestions) {
      if (question.id) {
        await managementApi.updateQuestion(targetQuizId, question.id, buildQuestionUpdatePayload(question))
      } else {
        await managementApi.createQuestion(targetQuizId, buildQuestionCreatePayload(question))
      }
    }

    toast.add({
      group: 'global',
      severity: 'success',
      summary: 'Квиз сохранен',
      detail: 'Изменения успешно сохранены.',
      life: 2500,
    })

    await router.push('/quizzes')
  } catch (error: unknown) {
    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Не удалось сохранить квиз',
      detail: getErrorMessage(error, 'Попробуйте снова.'),
      life: 3500,
    })
  } finally {
    isSaving.value = false
  }
}

onMounted(loadQuiz)

useHead({
  title: 'Редактор квиза',
})
</script>

<template>
  <section class="mx-auto flex w-full max-w-(--app-card-wide) flex-col gap-(--app-page-gap)">
    <header class="flex flex-col items-stretch justify-between gap-4 md:flex-row md:items-center">
      <div class="flex items-center gap-2">
        <Button icon="pi pi-arrow-left" text rounded aria-label="Назад" @click="handleCancel" />
      </div>

      <Button
        label="Сохранить"
        icon="pi pi-check"
        severity="success"
        :disabled="isLoading || isSaving"
        :loading="isSaving"
        @click="handleSave"
      />
    </header>

    <Card>
      <template #content>
        <div class="flex flex-col gap-4">
          <div class="flex flex-col gap-1.5">
            <label for="quiz_title" class="text-sm font-semibold">Название</label>
            <InputText id="quiz_title" v-model="quiz.title" placeholder="Введите название квиза" class="w-full" />
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="quiz_description" class="text-sm font-semibold">Описание</label>
            <Textarea id="quiz_description" v-model="quiz.description" rows="3" class="w-full" />
          </div>
        </div>
      </template>
    </Card>

    <VueDraggable
      v-model="quiz.questions"
      class="flex flex-col gap-4"
      handle=".question-card__drag-handle"
      :animation="150"
      @end="handleQuestionDragEnd"
    >
      <Card v-for="(question, qIndex) in quiz.questions" :key="question.localId">
        <template #content>
          <div class="flex flex-col gap-3">
            <div class="flex flex-col items-stretch justify-between gap-3 md:flex-row md:items-center">
              <div class="flex flex-wrap items-center gap-2">
                <i class="pi pi-bars question-card__drag-handle" aria-hidden="true" />
                <span class="text-sm font-bold text-(--app-color-text-muted)">#{{ qIndex + 1 }}</span>
                <SelectButton
                  v-model="question.selection_type"
                  :options="selectionOptions"
                  option-label="label"
                  option-value="value"
                  @change="switchQuestionType(question)"
                />
                <div class="inline-flex items-center gap-2 rounded-(--app-control-radius) bg-(--p-content-hover-background) px-2 py-1">
                  <i class="pi pi-clock" />
                  <InputNumber v-model="question.time_limit_seconds" :min="5" :max="120" suffix=" сек" />
                </div>
              </div>

              <div class="flex flex-wrap items-center gap-2">
                <Button
                  icon="pi pi-trash"
                  text
                  rounded
                  severity="danger"
                  aria-label="Удалить вопрос"
                  @click="removeQuestion(question.localId)"
                />
              </div>
            </div>

            <InputText v-model="question.text" placeholder="Текст вопроса" class="w-full" />

            <Divider align="left">
              <span class="text-[0.7rem] tracking-[0.08em] text-(--app-color-text-muted) uppercase">Варианты ответов</span>
            </Divider>

            <VueDraggable
              v-model="question.options"
              class="flex flex-col gap-2"
              handle=".question-card__option-handle"
              :animation="150"
              @end="handleOptionDragEnd(question)"
            >
              <div
                v-for="option in question.options"
                :key="option.localId"
                class="flex items-center gap-2 rounded-(--app-control-radius) border border-(--app-color-border) p-2.5"
                :class="{ 'question-card__option--correct': option.is_correct }"
              >
                <i class="pi pi-bars question-card__option-handle" aria-hidden="true" />
                <RadioButton
                  v-if="question.selection_type === 'single'"
                  :model-value="question.options.find((item) => item.is_correct)"
                  :value="option"
                  @update:model-value="toggleCorrect(question, option)"
                />
                <Checkbox v-else v-model="option.is_correct" binary />
                <InputText v-model="option.text" placeholder="Вариант ответа" class="w-full" />
              </div>
            </VueDraggable>
          </div>
        </template>
      </Card>
    </VueDraggable>

    <Button label="Добавить вопрос" icon="pi pi-plus" outlined class="w-full border-dashed" @click="addQuestion" />

    <Dialog v-model:visible="showCancelDialog" modal header="Отменить изменения" :draggable="false">
      <p class="m-0">Есть несохраненные изменения. Выйти без сохранения?</p>
      <template #footer>
        <div class="flex justify-end gap-2">
          <Button label="Остаться" text @click="showCancelDialog = false" />
          <Button label="Выйти" severity="danger" @click="confirmCancel" />
        </div>
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.question-card__drag-handle,
.question-card__option-handle {
  cursor: grab;
  color: var(--app-color-text-muted);
}

.question-card__option--correct {
  border-color: color-mix(in srgb, var(--app-color-primary) 45%, var(--app-color-border));
  background-color: color-mix(in srgb, var(--app-color-primary) 10%, transparent);
}
</style>
