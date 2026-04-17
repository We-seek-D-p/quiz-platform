<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Checkbox from 'primevue/checkbox'
import Divider from 'primevue/divider'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import RadioButton from 'primevue/radiobutton'
import SelectButton from 'primevue/selectbutton'
import { VueDraggable } from 'vue-draggable-plus'
import {
  createQuestionRequest,
  createQuizRequest,
  deleteQuestionRequest,
  getQuizQuestionsRequest,
  getQuizRequest,
  parseManagementErrorMessage,
  parseQuestionPublicList,
  parseQuizPublic,
  updateQuestionRequest,
  updateQuizRequest,
} from '@/api/quiz'
import { useAuthStore } from '@/stores/auth'
import type {
  QuestionCreate,
  QuestionUpdate,
  QuizDraft,
  QuizUpdate,
  QuestionDraft,
  OptionDraft,
} from '@/types'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const managementRequestOptions = {
  getAccessToken: () => authStore.accessToken,
  refreshAccessToken: authStore.refreshAccessToken,
}

const quiz = ref<QuizDraft>({
  title: '',
  description: '',
  questions: [],
})
const isLoading = ref(false)
const isSaving = ref(false)
const initialQuestionIds = ref<Set<string>>(new Set())

const selectionOptions = [
  { label: 'Один ответ', value: 'single' },
  { label: 'Несколько', value: 'multiple' },
]

const generateLocalId = () => Math.random().toString(36).substring(2, 9)

const getEditorQuizId = (): string | null => {
  const quizId = route.query.quiz
  if (typeof quizId === 'string' && quizId.length > 0 && quizId !== 'new') {
    return quizId
  }

  return null
}

const toDraftQuestion = (question: Awaited<ReturnType<typeof parseQuestionPublicList>>[number]): QuestionDraft => {
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

const loadQuiz = async () => {
  const quizId = getEditorQuizId()

  if (!quizId) {
    addQuestion()
    return
  }

  isLoading.value = true

  try {
    const [quizResponse, questionsResponse] = await Promise.all([
      getQuizRequest(quizId, managementRequestOptions),
      getQuizQuestionsRequest(quizId, managementRequestOptions),
    ])

    if (!quizResponse.ok) {
      throw new Error(await parseManagementErrorMessage(quizResponse))
    }

    if (!questionsResponse.ok) {
      throw new Error(await parseManagementErrorMessage(questionsResponse))
    }

    const quizPayload = await parseQuizPublic(quizResponse)
    const questions = await parseQuestionPublicList(questionsResponse)

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
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Не удалось загрузить квиз'
    alert(message)
    await router.push('/quizzes')
  } finally {
    isLoading.value = false
  }
}

const addQuestion = () => {
  quiz.value.questions.push({
    localId: generateLocalId(),
    text: '',
    selection_type: 'single',
    time_limit_seconds: 20,
    order_index: quiz.value.questions.length,
    options: Array.from({ length: 4 }, (_, i) => ({
      localId: generateLocalId(),
      text: '',
      is_correct: false,
      order_index: i,
    })),
  })
}

const removeQuestion = (localId: string) => {
  quiz.value.questions = quiz.value.questions
    .filter((question) => question.localId !== localId)
    .map((question, index) => ({
      ...question,
      order_index: index,
    }))
}

const addOption = (question: QuestionDraft) => {
  question.options.push({
    localId: generateLocalId(),
    text: '',
    is_correct: false,
    order_index: question.options.length,
  })
}

const removeOption = (question: QuestionDraft, optionId: string) => {
  if (question.options.length > 2) {
    question.options = question.options
      .filter((option) => option.localId !== optionId)
      .map((option, index) => ({
        ...option,
        order_index: index,
      }))
  }
}

const toggleCorrect = (question: QuestionDraft, option: OptionDraft) => {
  if (question.selection_type === 'single') {
    question.options.forEach((opt) => {
      opt.is_correct = opt.localId === option.localId
    })
  } else {
    option.is_correct = !option.is_correct
  }
}

const normalizeQuestion = (question: QuestionDraft, orderIndex: number) => {
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

    if (question.options.length < 2) {
      return `В вопросе #${questionIndex + 1} должно быть минимум 2 варианта`
    }

    for (const [optionIndex, option] of question.options.entries()) {
      if (option.text.trim().length === 0) {
        return `Заполните вариант #${optionIndex + 1} в вопросе #${questionIndex + 1}`
      }
    }
  }

  return null
}

const buildQuestionCreatePayload = (question: QuestionDraft): QuestionCreate => ({
  text: question.text,
  selection_type: question.selection_type,
  time_limit_seconds: question.time_limit_seconds,
  order_index: question.order_index,
  options: question.options.map((option) => ({
    text: option.text,
    order_index: option.order_index,
    is_correct: option.is_correct,
  })),
})

const buildQuestionUpdatePayload = (question: QuestionDraft): QuestionUpdate => ({
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
})

const handleCancel = async () => {
  if (confirm('У вас есть несохраненные изменения. Вы уверены, что хотите выйти?')) {
    await router.push('/quizzes')
  }
}

const handleSave = async () => {
  const validationError = validateQuizDraft()
  if (validationError) {
    alert(validationError)
    return
  }

  isSaving.value = true

  const normalizedQuestions = quiz.value.questions.map((question, index) =>
    normalizeQuestion(question, index),
  )

  try {
    const quizPayload: QuizUpdate = {
      title: quiz.value.title.trim(),
      description: quiz.value.description.trim(),
    }

    let targetQuizId = quiz.value.id

    if (targetQuizId) {
      const response = await updateQuizRequest(targetQuizId, quizPayload, managementRequestOptions)

      if (!response.ok) {
        throw new Error(await parseManagementErrorMessage(response))
      }
    } else {
      const response = await createQuizRequest(
        {
          title: quizPayload.title ?? '',
          description: quizPayload.description ?? '',
        },
        managementRequestOptions,
      )

      if (!response.ok) {
        throw new Error(await parseManagementErrorMessage(response))
      }

      const createdQuiz = await parseQuizPublic(response)
      targetQuizId = createdQuiz.id
      quiz.value.id = createdQuiz.id
    }

    if (!targetQuizId) {
      throw new Error('Не удалось определить идентификатор квиза')
    }

    const currentExistingIds = new Set(
      normalizedQuestions
        .map((question) => question.id)
        .filter((questionId): questionId is string => Boolean(questionId)),
    )

    const deletedQuestionIds = [...initialQuestionIds.value].filter(
      (questionId) => !currentExistingIds.has(questionId),
    )

    for (const deletedQuestionId of deletedQuestionIds) {
      const deleteResponse = await deleteQuestionRequest(
        targetQuizId,
        deletedQuestionId,
        managementRequestOptions,
      )

      if (!deleteResponse.ok) {
        throw new Error(await parseManagementErrorMessage(deleteResponse))
      }
    }

    for (const question of normalizedQuestions) {
      if (question.id) {
        const updateResponse = await updateQuestionRequest(
          targetQuizId,
          question.id,
          buildQuestionUpdatePayload(question),
          managementRequestOptions,
        )

        if (!updateResponse.ok) {
          throw new Error(await parseManagementErrorMessage(updateResponse))
        }
      } else {
        const createResponse = await createQuestionRequest(
          targetQuizId,
          buildQuestionCreatePayload(question),
          managementRequestOptions,
        )

        if (!createResponse.ok) {
          throw new Error(await parseManagementErrorMessage(createResponse))
        }
      }
    }

    await router.push('/quizzes')
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Не удалось сохранить квиз'
    alert(message)
  } finally {
    isSaving.value = false
  }
}

onMounted(loadQuiz)
</script>

<template>
  <div class="max-w-4xl mx-auto p-4">
    <div class="flex justify-between items-center mb-8">
      <div class="flex items-center gap-4">
        <Button icon="pi pi-arrow-left" text rounded @click="handleCancel()" />
        <h1 class="text-2xl font-bold tracking-tight">Редактор квиза</h1>
      </div>
      <Button
        label="Сохранить"
        icon="pi pi-check"
        severity="success"
        :disabled="isLoading || isSaving"
        :loading="isSaving"
        @click="handleSave"
      />
    </div>

    <Card class="mb-6 border-none shadow-sm">
      <template #content>
        <div class="flex flex-col gap-4">
          <div class="flex flex-col gap-2">
            <label class="font-semibold opacity-70">Название</label>
            <InputText v-model="quiz.title" placeholder="Введите название" class="w-full p-3" />
          </div>
          <div class="flex flex-col gap-2">
            <InputText v-model="quiz.description" placeholder="Введите Описание квиза" class="w-full p-5 text-sm" />
          </div>
        </div>
      </template>
    </Card>

    <VueDraggable
      v-model="quiz.questions"
      :animation="150"
      handle=".drag-handle"
      class="space-y-6"
    >
      <Card v-for="(question, qIndex) in quiz.questions" :key="question.localId" class="border-none shadow-sm">
        <template #content>
          <div class="flex flex-col gap-4">
            <div class="flex justify-between items-center">
              <div class="flex items-center gap-3">
                <i class="pi pi-bars drag-handle cursor-grab opacity-30 hover:opacity-100 transition-opacity p-2" />
                <span class="text-sm font-bold opacity-30 uppercase">#{{ qIndex + 1 }}</span>
                <SelectButton
                  v-model="question.selection_type"
                  :options="selectionOptions"
                  option-label="label"
                  option-value="value"
                  class="scale-75 origin-left"
                  @change="question.options.forEach(o => o.is_correct = false)"
                />
                <div class="timer-badge">
                  <i class="pi pi-clock opacity-50 text-[10px]" />
                  <InputNumber
                    v-model="question.time_limit_seconds"
                    suffix=" сек"
                    :min="5"
                    :max="60"
                    input-class="text-sm font-bold focus:ring-0 bg-transparent border-none w-20"
                  />
                </div>
              </div>
              <Button icon="pi pi-trash" severity="danger" text rounded @click="removeQuestion(question.localId)" />
            </div>

            <InputText v-model="question.text" placeholder="Текст вопроса" class="w-full p-3 border-gray-200" />

            <Divider align="left">
              <span class="text-[10px] opacity-40 font-bold uppercase tracking-widest">Варианты ответов</span>
            </Divider>

            <div class="flex flex-col gap-3">
              <VueDraggable
                v-model="question.options"
                handle=".option-handle"
                :animation="150"
                class="flex flex-col gap-3"
              >
                <div
                  v-for="option in question.options"
                  :key="option.localId"
                  class="flex items-center gap-3 p-3 border rounded-xl transition-all duration-200 option-box group"
                  :class="{ 'correct-option': option.is_correct }"
                >
                  <i class="pi pi-bars option-handle cursor-grab opacity-30 hover:opacity-100 p-1" />
                  <RadioButton
                    v-if="question.selection_type === 'single'"
                    :model-value="question.options.find(o => o.is_correct)"
                    :value="option"
                    severity="success"
                    @update:model-value="toggleCorrect(question, option)"
                  />
                  <Checkbox
                    v-else
                    v-model="option.is_correct"
                    binary
                    severity="success"
                  />
                  <InputText v-model="option.text" placeholder="Ответ" class="w-full border-none bg-transparent shadow-none p-0 focus:ring-0" />

                  <Button
                    v-if="question.options.length > 2"
                    icon="pi pi-times"
                    severity="danger"
                    text
                    rounded
                    class="opacity-0 group-hover:opacity-100 transition-opacity w-8 h-8 p-0"
                    @click="removeOption(question, option.localId)"
                  />
                </div>
              </VueDraggable>
              <button
                class="mt-2 flex items-center justify-center gap-2 p-3 border border-dashed rounded-xl opacity-50 hover:opacity-100 transition-all text-sm font-medium w-full"
                @click="addOption(question)"
              >
                <i class="pi pi-plus text-xs" />
                Добавить вариант
              </button>
            </div>
          </div>
        </template>
      </Card>
    </VueDraggable>
    <Button
      label="Добавить вопрос"
      icon="pi pi-plus"
      outlined
      class="w-full py-4 border-dashed bg-transparent mt-6"
      @click="addQuestion"
    />
  </div>
</template>

<style scoped>
.option-box {
  border-color: var(--p-content-border-color);
  background: var(--p-content-background);
}

.correct-option {
  background-color: rgba(77, 222, 128);
}

.timer-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.1rem 0.6rem;
  border-radius: 0.5rem;
  background-color: var(--p-content-hover-background);
}
</style>
