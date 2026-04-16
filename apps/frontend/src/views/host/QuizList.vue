<script setup lang="ts">
import { ref, onMounted } from 'vue';
import Button from 'primevue/button'
import type { QuizTransport } from '@/types';


const quizzes = ref<QuizTransport[]>([]);


const fetchQuizzes = async () => {
quizzes.value = [
    { id: '1', title: 'Тестовый квиз', createdAt: new Date().toISOString(), questions: [] }
    ] as any;
};

const handleDelete = async (id: string) => {
  if (confirm('Вы уверены, что хотите удалить этот квиз?')) {
    // await api.delete(`/quizzes/${id}`);
    quizzes.value = quizzes.value.filter(q => q.id !== id);
  }
};

onMounted(fetchQuizzes);
</script>


<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold">Мои квизы</h1>
      <Button
            label="Новый квиз"
            @click="$router.push('/host/quizzes/new')"
          />
    </div>

    <div class="bg-white border rounded-xl overflow-hidden">
      <table class="w-full text-left">
        <thead class="bg-gray-50 border-b">
          <tr>
            <th class="p-4 font-semibold">Название</th>
            <th class="p-4 font-semibold">Вопросов</th>
            <th class="p-4 font-semibold">Дата создания</th>
            <th class="p-4 text-right">Действия</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="quiz in quizzes" :key="quiz.id" class="border-b last:border-0 hover:bg-gray-50">
            <td class="p-4 font-medium">{{ quiz.title }}</td>
            <td class="p-4 text-gray-600">{{ quiz.questions?.length || 0 }}</td>
            <td class="p-4 text-gray-500 text-sm">
              {{ new Date(quiz.createdAt).toLocaleDateString() }}
            </td>
            <td class="p-4 text-right space-x-2">
              <button 
                @click="$router.push(`/host/quizzes/${quiz.id}/edit`)"
                class="text-blue-600 hover:underline"
              >
                Изменить
              </button>
              <button 
                @click="handleDelete(quiz.id)"
                class="text-red-600 hover:underline"
              >
                Удалить
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="quizzes.length === 0" class="p-10 text-center text-gray-500">
        У вас пока нет квизов.
      </div>
    </div>
  </div>
</template>
