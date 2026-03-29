<script setup lang="ts"></script>

<template>
  <div class="flex flex-col min-h-screen bg-transparent">
    <main class="flex flex-1 items-center justify-center p-2 sm:p-4">
      <div class="w-full max-w-4xl">
        <slot />
      </div>
    </main>
  </div>
</template>

<style scoped></style>
