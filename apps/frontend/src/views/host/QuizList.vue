<script setup lang="ts">
import { ref, onMounted } from 'vue';
import Button from 'primevue/button';
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import type { QuizTransport } from '@/types';

const quizzes = ref<QuizTransport[]>([]);

const fetchQuizzes = async () => {
  quizzes.value = [
    { id: '0', title: 'Test quiz', createdAt: new Date().toISOString(), questions: [] }
  ] as any;
};

const handleDelete = async (id: string) => {
  if (confirm('Вы уверены, что хотите удалить этот квиз?')) {
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
        icon="pi pi-plus"
        @click="$router.push('/quizzes/editor')"
      />
    </div>
    
    <DataTable :value="quizzes" class=" rounded-xl overflow-hidden">
      <Column field="title" header="Название"></Column>
      
      <Column header="Вопросов">
        <template #body="slotProps">
          {{ slotProps.data.questions?.length || 0 }}
        </template>
      </Column>

      <Column header="Дата создания">
        <template #body="slotProps">
          {{ new Date(slotProps.data.createdAt).toLocaleDateString() }}
        </template>
      </Column>

      <Column headerStyle="text-align: right" bodyStyle="text-align: right">
        <template #body="slotProps">
          <div class="flex justify-end gap-2">
            <Button
              icon="pi pi-play" 
              text 
              style="color: green"
              @click="$router.push('/lobby')"
            />
            <Button 
              icon="pi pi-pencil" 
              text 
              severity="secondary" 
              @click="$router.push(`/quizzes/editor?quiz=${slotProps.data.id}`)" 
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
