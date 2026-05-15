import type { SessionPhase } from '~/types/session-ws'

interface UsePhaseTimerOptions {
  phase: Ref<SessionPhase>
  deadlineAt: Ref<string | null>
  revealUntil: Ref<string | null>
  questionTimeLimitSeconds: Ref<number | null>
  revealDurationSec: Ref<number>
}

export function usePhaseTimer(options: UsePhaseTimerOptions) {
  const timerProgress = ref(0)
  const timerLabel = ref('--')
  let timerInterval: ReturnType<typeof setInterval> | null = null
  let currentPhaseTotalMs = 1000

  const clearTimer = () => {
    if (timerInterval) {
      clearInterval(timerInterval)
      timerInterval = null
    }
  }

  const startTimer = () => {
    clearTimer()
    const target = options.phase.value === 'question_open'
      ? options.deadlineAt.value
      : options.phase.value === 'answer_reveal'
        ? options.revealUntil.value
        : null

    if (target) {
      if (options.phase.value === 'question_open' && options.questionTimeLimitSeconds.value) {
        currentPhaseTotalMs = options.questionTimeLimitSeconds.value * 1000
      } else if (options.phase.value === 'answer_reveal' && options.revealDurationSec.value) {
        currentPhaseTotalMs = options.revealDurationSec.value * 1000
      } else {
        const remaining = new Date(target).getTime() - Date.now()
        currentPhaseTotalMs = Math.max(remaining, 1000)
      }
    } else {
      currentPhaseTotalMs = 1000
    }

    recomputeTimer()
    timerInterval = setInterval(recomputeTimer, 50)
  }

  const recomputeTimer = () => {
    const target = options.phase.value === 'question_open'
      ? options.deadlineAt.value
      : options.phase.value === 'answer_reveal'
        ? options.revealUntil.value
        : null

    if (!target) {
      timerProgress.value = 0
      timerLabel.value = '--'
      return
    }

    const endMs = new Date(target).getTime()
    const nowMs = Date.now()
    const remainingMs = endMs - nowMs

    if (remainingMs <= 0) {
      timerLabel.value = '0s'
      timerProgress.value = 0
      return
    }

    const remainingSec = Math.ceil(remainingMs / 1000)
    timerLabel.value = `${remainingSec}s`

    const progress = (remainingMs / currentPhaseTotalMs) * 100
    timerProgress.value = Math.max(0, Math.min(100, progress))
  }

  watch(
    () => [
      options.phase.value,
      options.deadlineAt.value,
      options.revealUntil.value,
      options.questionTimeLimitSeconds.value,
      options.revealDurationSec.value,
    ],
    () => {
      startTimer()
    },
    { deep: true }
  )

  onMounted(startTimer)
  onBeforeUnmount(clearTimer)

  return {
    timerLabel,
    timerProgress,
  }
}
