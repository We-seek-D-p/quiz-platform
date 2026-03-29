<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import Password from 'primevue/password'
import InputText from 'primevue/inputtext'
import FloatLabel from 'primevue/floatlabel'


const SendLogin = (): void => {}
</script>

<template>
  <main class="app-root">
    <Card class="demo-card p-anchored-overlay-enter-active">
      <template #title>Login</template>
      <template #content>
        <FloatLabel variant="in" class="m-2">
          <InputText id="on_label" v-model="value" autocomplete="off" />
          <label for="on_label">Email</label>
        </FloatLabel>

        <FloatLabel variant="in" class="m-2">
          <Password v-model="value" :feedback="false" />
          <label for="on_label">Password</label>
        </FloatLabel>

        <div class="flex justify-center w-full mt-3">
          <Button label="Login" @click="SendLogin" />
        </div>
      </template>

      <template #footer>
        <span class="text-xs">Doesn't have an account yet? </span>
        <router-link to="/register"
          ><span class="text-xs font-bold text-primary cursor-pointer select-none"
            >Register</span
          ></router-link
        >
      </template>
    </Card>
  </main>
</template>

<style scoped>
.app-root {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 1rem;
}

.demo-card {
  width: min(28rem, 100%);
}

.count-text {
  margin: 0;
}

.theme-button {
  width: min(28rem, 100%);
}
</style>
