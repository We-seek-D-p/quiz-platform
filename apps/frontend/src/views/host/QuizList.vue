<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import type { QuizPublic } from '@/types'
import { useAuthStore } from '@/stores/auth'
import {
  deleteQuizRequest,
  getQuizQuestionsRequest,
  getQuizzesRequest,
  parseManagementErrorMessage,
  parseQuestionPublicList,
  parseQuizPublicList,
} from '@/api/quiz'

type QuizTableRow = QuizPublic & {
  questionCount: number
}

const router = useRouter()
const authStore = useAuthStore()
const managementRequestOptions = {
  getAccessToken: () => authStore.accessToken,
  refreshAccessToken: authStore.refreshAccessToken,
}

const quizzes = ref<QuizTableRow[]>([])
const isLoading = ref(false)

const fetchQuizzes = async () => {
  isLoading.value = true

  try {
    const response = await getQuizzesRequest(managementRequestOptions)
    if (!response.ok) {
      throw new Error(await parseManagementErrorMessage(response))
    }

    const quizList = await parseQuizPublicList(response)
    const questionCounts = await Promise.all(
      quizList.map(async (quiz) => {
        const questionsResponse = await getQuizQuestionsRequest(quiz.id, {
          ...managementRequestOptions,
        })

        if (!questionsResponse.ok) {
          return 0
        }

        const questions = await parseQuestionPublicList(questionsResponse)
        return questions.length
      }),
    )

    quizzes.value = quizList.map((quiz, index) => ({
      ...quiz,
      questionCount: questionCounts[index] ?? 0,
    }))
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Не удалось загрузить квизы'
    alert(message)
  } finally {
    isLoading.value = false
  }
}

const handleDelete = async (id: string) => {
  if (!confirm('Вы уверены, что хотите удалить этот квиз?')) {
    return
  }

  try {
    const response = await deleteQuizRequest(id, managementRequestOptions)
    if (!response.ok) {
      throw new Error(await parseManagementErrorMessage(response))
    }

    quizzes.value = quizzes.value.filter((quiz) => quiz.id !== id)
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Не удалось удалить квиз'
    alert(message)
  }
}

const handleEdit = (quizId: string) => {
  router.push(`/quizzes/editor?quiz=${quizId}`)
}

onMounted(fetchQuizzes)
</script>

<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold">Мои квизы</h1>
      <Button
        label="Новый квиз"
        icon="pi pi-plus"
        @click="$router.push('/quizzes/editor')"
      />
    </div>
    
    <DataTable :value="quizzes" class=" rounded-xl overflow-hidden" :loading="isLoading">
      <Column field="title" header="Название"></Column>
      
      <Column header="Вопросов">
        <template #body="slotProps">
          {{ slotProps.data.questionCount }}
        </template>
      </Column>

      <Column header="Дата создания">
        <template #body="slotProps">
          {{ new Date(slotProps.data.created_at).toLocaleDateString() }}
        </template>
      </Column>

      <Column header-style="text-align: right" body-style="text-align: right">
        <template #body="slotProps">
          <div class="flex justify-end gap-2">
            <Button
              icon="pi pi-play" 
              text 
              style="color: green"
              disabled
              />
             <Button 
               icon="pi pi-pencil" 
               text 
               severity="secondary" 
               @click="handleEdit(slotProps.data.id)" 
             />
            <Button 
              icon="pi pi-trash" 
              text 
              severity="danger" 
              @click="handleDelete(slotProps.data.id)" 
            />
          </div>
        </template>
      </Column>

      <template #empty>
        <div class="p-4 text-center">У вас еще нет квизов.</div>
      </template>
    </DataTable>
  </div>
</template>
