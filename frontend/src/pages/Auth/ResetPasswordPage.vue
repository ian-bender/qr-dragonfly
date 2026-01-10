<script setup lang="ts">
import { ref, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usersApi } from '../../api'
import { authErrorMessage } from '../../lib/authErrors'
import AppButton from '../../components/ui/AppButton.vue'

const route = useRoute()
const router = useRouter()

const email = ref('')
const code = ref('')
const newPassword = ref('')

const busy = ref(false)
const errorMessage = ref<string | null>(null)
const statusMessage = ref<string | null>(null)

watchEffect(() => {
  const q = route.query.email
  if (typeof q === 'string' && !email.value) email.value = q
})

async function submit() {
  errorMessage.value = null
  statusMessage.value = null

  const e = email.value.trim().toLowerCase()
  const c = code.value.trim()
  const p = newPassword.value.trim()
  if (!e || !c || !p) {
    errorMessage.value = 'Email, code, and new password are required.'
    return
  }

  busy.value = true
  try {
    await usersApi.confirmForgotPassword({ email: e, code: c, newPassword: p })
    statusMessage.value = 'Password reset. You can now log in.'
    await router.push({ name: 'login', query: { email: e } })
  } catch (err) {
    errorMessage.value = authErrorMessage(err)
  } finally {
    busy.value = false
    newPassword.value = ''
  }
}
</script>

<template>
  <main class="page">
    <header class="header">
      <h1 class="title">Reset password</h1>
      <p class="subtitle">Enter the reset code from your email.</p>
    </header>

    <section class="card">
      <h2 class="sectionTitle">Reset</h2>

      <form class="form" @submit.prevent="submit">
        <label class="field">
          <span class="label">Email</span>
          <input v-model="email" class="input" type="email" autocomplete="email" />
        </label>

        <label class="field">
          <span class="label">Reset code</span>
          <input v-model="code" class="input" type="text" inputmode="numeric" autocomplete="one-time-code" />
        </label>

        <label class="field">
          <span class="label">New password</span>
          <input v-model="newPassword" class="input" type="password" autocomplete="new-password" />
        </label>

        <div class="actions">
          <AppButton type="submit" :disabled="busy">{{ busy ? 'Resetting…' : 'Reset password' }}</AppButton>
          <RouterLink class="muted" to="/forgot-password">Need a code?</RouterLink>
        </div>
      </form>

      <p v-if="statusMessage" class="status">{{ statusMessage }}</p>
      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

      <div class="links">
        <AppButton to="/login" variant="secondary">Go To Login</AppButton>
      </div>
    </section>
  </main>
</template>

<style scoped src="./AuthPage.scss" lang="scss"></style>
