<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import Button from 'primevue/button';
import InputText from 'primevue/inputtext';
import InputNumber from 'primevue/inputnumber';
import Card from 'primevue/card';
import Divider from 'primevue/divider';
import RadioButton from 'primevue/radiobutton';
import { VueDraggable } from 'vue-draggable-plus'
import Checkbox from 'primevue/checkbox';
import SelectButton from 'primevue/selectbutton';
import type { QuizDraft, QuestionDraft, OptionDraft } from '@/types/quiz-editor';

const route = useRoute();
const router = useRouter();

const quiz = ref<QuizDraft>({
  title: '',
  description: '',
  questions: []
});

const selectionOptions = [
  { label: 'Один ответ', value: 'single' },
  { label: 'Несколько', value: 'multiple' }
];

const generateLocalId = () => Math.random().toString(36).substring(2, 9);

onMounted(() => {
  const quizId = route.query.quiz;
  if (quizId && quizId !== 'new') {
    quiz.value = {
      id: quizId as string,
      title: 'Test quiz',
      description: 'description',
      questions: [
        {
          localId: generateLocalId(),
          text: 'Question title',
          selection_type: 'single',
          time_limit_seconds: 30,
          order_index: 0,
          options: Array.from({ length: 4 }, (_, i) => ({
            localId: generateLocalId(),
            text: `Option ${i + 1}`,
            is_correct: i === 0,
            order_index: i
          }))
        }
      ]
    };
  } else {
    addQuestion();
  }
});

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
      order_index: i
    }))
  });
};

const removeQuestion = (localId: string) => {
  quiz.value.questions = quiz.value.questions.filter(q => q.localId !== localId);
};

const addOption = (question: QuestionDraft) => {
  question.options.push({
    localId: generateLocalId(),
    text: '',
    is_correct: false,
    order_index: question.options.length
  });
};

const removeOption = (question: QuestionDraft, optionId: string) => {
  if (question.options.length > 2) {
    question.options = question.options.filter(o => o.localId !== optionId);
  }
};

const toggleCorrect = (question: QuestionDraft, option: OptionDraft) => {
  if (question.selection_type === 'single') {
    question.options.forEach(opt => opt.is_correct = opt.localId === option.localId);
  } else {
    option.is_correct = !option.is_correct;
  }
};

const handleCancel = async () => {
  if (confirm('У вас есть несохраненные изменения. Вы уверены, что хотите выйти?')) {
    router.push('/quizzes');
  } else {
    console.log("Stays on page")
  }
};

const handleSave = () => {
  router.push('/quizzes');
};
// todo: functionality
</script>

<template>
  <div class="max-w-4xl mx-auto p-4">
    <div class="flex justify-between items-center mb-8">
      <div class="flex items-center gap-4">
        <Button icon="pi pi-arrow-left" text rounded @click="handleCancel()" />
        <h1 class="text-2xl font-bold tracking-tight">Редактор квиза</h1>
      </div>
      <Button label="Сохранить" icon="pi pi-check" severity="success" @click="handleSave" />
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
                  optionLabel="label" 
                  optionValue="value" 
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
                    inputClass="text-sm font-bold focus:ring-0 bg-transparent border-none w-20"
                  />
                </div>
              </div>
              <Button icon="pi pi-trash" severity="danger" text rounded @click="removeQuestion(question.localId)" />
            </div>

            <InputText v-model="question.text" placeholder="Текст вопроса" class="w-full p-3 border-gray-200" />

            <Divider align="left">
              <span class="text-[10px] opacity-40 font-bold uppercase tracking-widest">Варианты ответов</span>
            </Divider>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div 
                v-for="option in question.options" :key="option.localId" 
                class="flex items-center gap-3 p-3 border rounded-xl transition-all duration-200 option-box group"
                :class="{ 'correct-option': option.is_correct }"
              >
                <RadioButton 
                  v-if="question.selection_type === 'single'"
                  :modelValue="question.options.find(o => o.is_correct)" 
                  :value="option" 
                  @update:modelValue="toggleCorrect(question, option)"
                  severity="success"
                />
                <Checkbox 
                  v-else
                  v-model="option.is_correct" 
                  binary
                  severity="success"
                />
                <InputText v-model="option.text" placeholder="Ответ" class="w-full border-none bg-transparent shadow-none p-0 focus:ring-0" />
                <Button 
                  icon="pi pi-times" 
                  severity="danger" 
                  text 
                  rounded 
                  class="opacity-0 group-hover:opacity-100 transition-opacity w-8 h-8 p-0"
                  @click="removeOption(question, option.localId)"
                  v-if="question.options.length > 2"
                />
              </div>
              <button 
                @click="addOption(question)"
                class="flex items-center justify-center gap-2 p-3 border border-dashed rounded-xl opacity-50 hover:opacity-100 transition-all text-sm font-medium"
              >
                <i class="pi pi-plus text-xs" />
                Добавить вариант
              </button>
            </div>
          </div>
        </template>
      </Card>
      <Button label="Добавить вопрос" icon="pi pi-plus" outlined class="w-full py-4 border-dashed bg-transparent" @click="addQuestion" />
    </VueDraggable>
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
