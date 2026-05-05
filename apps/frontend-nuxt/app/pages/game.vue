<script setup lang="ts">
import Card from 'primevue/card'
import ProgressBar from 'primevue/progressbar'
import Button from 'primevue/button'

const currentQuestionIndex = ref(0)
const timerProgress = ref(80)
const isSubmitted = ref(false)

const testQuestion = {
  text: 'В каком году был основан МАИ?',
  selection_type: 'single', // multiple
  options: [
    { id: 1, text: '1928', is_correct: false },
    { id: 2, text: '1930', is_correct: true },
    { id: 3, text: '1942', is_correct: false },
    { id: 4, text: '1954', is_correct: false }
  ]
}

const selectedIds = ref<number[]>([])

const toggleOption = (id: number) => {
  if (isSubmitted.value) return
  selectedIds.value = [id]
}

const handleAnswer = () => {
  isSubmitted.value = true
}

const getOptionClass = (option: any) => {
  return {
    'option-btn--active': selectedIds.value.includes(option.id) && !isSubmitted.value,
    'option-btn--correct': isSubmitted.value && option.is_correct,
    'option-btn--wrong': isSubmitted.value && selectedIds.value.includes(option.id) && !option.is_correct
  }
}
</script>

<template>
    <div class="quiz-game__container">
      <Card class="quiz-card">
        <template #content>
          <ProgressBar :value="timerProgress" :show-value="false" class="quiz-game__progress" />
          <div class="quiz-game__main">
            <div class="quiz-game__question-wrap">
              <span class="quiz-game__meta">Вопрос {{ currentQuestionIndex + 1 }}</span>
              <h1 class="quiz-game__title">{{ testQuestion.text }}</h1>
            </div>

            <div class="quiz-game__options">
              <button
                v-for="option in testQuestion.options"
                :key="option.id"
                class="option-btn"
                :class="getOptionClass(option)"
                @click="toggleOption(option.id)"
              >
                <span class="option-btn__text">{{ option.text }}</span>
                <i v-if="selectedIds.includes(option.id) && !isSubmitted" class="pi pi-check" />
              </button>
            </div>

            <div class="quiz-game__actions">
              <Button
                :label="'Ответить'"
                :disabled="selectedIds.length === 0"
                class="quiz-game__submit"
                @click="handleAnswer"
              />
            </div>
          </div>
        </template>
      </Card>
    </div>
</template>

<style scoped>
.quiz-game {
  display: flex;
  flex-direction: column;
  height: 100vh;
  height: 100dvh;
  background-color: var(--p-content-hover-background);
}

.quiz-game__header {
  padding: 1.5rem 2rem;
  width: 100%;
}

.quiz-game__progress {
  height: 8px;
}

.quiz-game__container {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 2rem;
}

.quiz-card {
  width: 100%;
  max-width: 700px;
  border-radius: 1.5rem;
}

.quiz-game__main {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2.5rem;
  padding: 1rem;
}

.quiz-game__question-wrap {
  text-align: center;
}

.quiz-game__meta {
  font-size: 0.875rem;
  font-weight: 800;
  text-transform: uppercase;
  color: var(--app-color-text-muted);
}

.quiz-game__title {
  margin: 0.75rem 0 0;
  font-size: 2rem;
  font-weight: 700;
}

.quiz-game__options {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  width: 100%;
}

.option-btn {
  background: var(--p-content-background);
  border: 1px solid var(--app-color-border);
  border-radius: 1rem;
  padding: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  transition: all 0.2s;
  min-height: 70px;
}

.option-btn--active {
  border-color: var(--app-color-primary);
  background: color-mix(in srgb, var(--app-color-primary) 8%, transparent);
}

.option-btn--correct {
  border-color: rgb(34, 197, 94);
  background: rgba(34, 197, 94, 0.2);
}

.option-btn--wrong {
  border-color: rgba(255, 98, 98);
  background: rgba(255, 98, 98, 0.2)
}

.option-btn__text {
  font-size: 1.1rem;
  font-weight: 700;
}

.quiz-game__actions {
  width: 100%;
  display: flex;
  justify-content: center;
}

.quiz-game__submit {
  width: 100%;
  max-width: 300px;
  padding: 1rem;
  font-weight: 800;
}

@media (max-width: 768px) {
  .quiz-game__options {
    grid-template-columns: 1fr;
  }
}
</style>
