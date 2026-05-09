<template>
  <div class="game-layout">
    <main class="game-layout__main">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.game-layout {
  min-height: 100dvh;
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--app-color-primary) 10%, transparent), transparent 48%),
    radial-gradient(circle at bottom left, color-mix(in srgb, var(--app-color-primary-hover) 10%, transparent), transparent 52%),
    var(--app-color-bg);
}

.game-layout__main {
  min-height: 100dvh;
  padding: clamp(0.75rem, 2vw, 1.5rem);
}
</style>
