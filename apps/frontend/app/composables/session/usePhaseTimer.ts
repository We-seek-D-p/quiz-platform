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
    if (timerInterval) {
      clearInterval(timerInterval)
      timerInterval = null
    }
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

    if (options.phase.value === 'question_open') {
      const totalSec = options.questionTimeLimitSeconds.value || 1
      const progress = (remainingMs / (totalSec * 1000)) * 100
      timerProgress.value = Math.max(0, Math.min(100, progress))
    } else if (options.phase.value === 'answer_reveal') {
      const revealMs = (options.revealDurationSec.value || 1) * 1000
      const progress = (remainingMs / revealMs) * 100
      timerProgress.value = Math.max(0, Math.min(100, progress))
    } else {
      timerProgress.value = 0
    }
  }

  const startTimer = () => {
    clearTimer()
    recomputeTimer()
    timerInterval = setInterval(recomputeTimer, 100)
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
