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

  const clearTimer = () => {
    if (!timerInterval) {
      return
    }
    clearInterval(timerInterval)
    timerInterval = null
  }

  const recomputeTimer = () => {
    const countdownTarget =
      options.phase.value === 'question_open'
        ? options.deadlineAt.value
        : options.phase.value === 'answer_reveal'
          ? options.revealUntil.value
          : null

    if (!countdownTarget) {
      timerProgress.value = 0
      timerLabel.value = '--'
      return
    }

    const endMs = new Date(countdownTarget).getTime()
    const nowMs = Date.now()
    const remainingMs = Math.max(0, endMs - nowMs)
    const remainingSec = Math.ceil(remainingMs / 1000)
    timerLabel.value = `${remainingSec}s`

    if (options.phase.value === 'question_open' && options.questionTimeLimitSeconds.value) {
      const total = Math.max(1, options.questionTimeLimitSeconds.value)
      timerProgress.value = Math.min(100, Math.max(0, (remainingSec / total) * 100))
      return
    }

    if (options.phase.value === 'answer_reveal') {
      const revealWindowMs = Math.max(1, options.revealDurationSec.value) * 1000
      timerProgress.value = Math.min(100, Math.max(0, (remainingMs / revealWindowMs) * 100))
      return
    }

    timerProgress.value = 0
  }

  const startTimer = () => {
    clearTimer()
    recomputeTimer()
    timerInterval = setInterval(recomputeTimer, 300)
  }

  watch(
    () =>
      [
        options.phase.value,
        options.deadlineAt.value,
        options.revealUntil.value,
        options.questionTimeLimitSeconds.value,
        options.revealDurationSec.value,
      ] as const,
    () => {
      startTimer()
    },
  )

  onMounted(startTimer)
  onBeforeUnmount(clearTimer)

  return {
    timerLabel,
    timerProgress,
  }
}
