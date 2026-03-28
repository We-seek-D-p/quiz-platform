<script setup lang="ts">
import {useRouter} from 'vue-router';
import Button from 'primevue/button';
import ThemeSwitcher from '@/components/common/ThemeSwitcher.vue';

const router = useRouter();
</script>

<template>
  <div class="min-h-screen flex flex-col bg-transparent">

    <header
        class="h-14 flex justify-between items-center px-4 border-b border-surface-200 dark:border-surface-800 bg-surface-0 dark:bg-surface-900">

      <div class="text-xl font-bold text-primary cursor-pointer select-none" @click="router.push('/')">
        Quiz Platform
      </div>

      <div class="flex items-center gap-3">
        <ThemeSwitcher/>

        <Button
            icon="pi pi-user"
            variant="outlined"
            class="layout-icon-button"
            @click="router.push('/host')"
        />
      </div>
    </header>

    <main class="flex-1 flex items-center justify-center p-6">
      <slot/>
    </main>
  </div>
</template>
