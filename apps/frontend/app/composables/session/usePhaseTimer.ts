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

  const getTimerTarget = () => {
    if (options.phase.value === 'question_open') {
      return options.deadlineAt.value
    }

    if (options.phase.value === 'answer_reveal') {
      return options.revealUntil.value
    }

    return null
  }

  const getPhaseTotalMs = (endMs: number) => {
    if (options.phase.value === 'question_open' && options.questionTimeLimitSeconds.value) {
      return options.questionTimeLimitSeconds.value * 1000
    }

    if (options.phase.value === 'answer_reveal' && options.revealDurationSec.value) {
      return options.revealDurationSec.value * 1000
    }

    return Math.max(endMs - Date.now(), 1000)
  }

  const setIdleTimer = () => {
    timerProgress.value = 0
    timerLabel.value = '--'
  }

  const setTimerValues = (endMs: number, forceMax = false) => {
    const remainingMs = endMs - Date.now()

    if (remainingMs <= 0) {
      timerLabel.value = '0s'
      timerProgress.value = 0
      return
    }

    timerLabel.value = `${Math.ceil(remainingMs / 1000)}s`

    const progress = (remainingMs / currentPhaseTotalMs) * 100
    const clampedProgress = Math.max(0, Math.min(100, progress))
    timerProgress.value = forceMax && clampedProgress >= 95 ? 100 : clampedProgress
  }

  const startTimer = () => {
    clearTimer()
    const target = getTimerTarget()

    if (!target) {
      currentPhaseTotalMs = 1000
      setIdleTimer()
      return
    }

    const endMs = new Date(target).getTime()
    if (!Number.isFinite(endMs)) {
      currentPhaseTotalMs = 1000
      setIdleTimer()
      return
    }

    currentPhaseTotalMs = Math.max(getPhaseTotalMs(endMs), 1000)
    setTimerValues(endMs, true)
    timerInterval = setInterval(recomputeTimer, 50)
  }

  const recomputeTimer = () => {
    const target = getTimerTarget()

    if (!target) {
      setIdleTimer()
      return
    }

    const endMs = new Date(target).getTime()
    if (!Number.isFinite(endMs)) {
      setIdleTimer()
      return
    }

    setTimerValues(endMs)
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
  )

  onMounted(startTimer)
  onBeforeUnmount(clearTimer)

  return {
    timerLabel,
    timerProgress,
  }
}
