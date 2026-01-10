<script setup lang="ts">
import { ref, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usersApi } from '../../api'
import { authErrorMessage } from '../../lib/authErrors'
import AppButton from '../../components/ui/AppButton.vue'

const route = useRoute()
const router = useRouter()

const email = ref('')

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
  if (!e) {
    errorMessage.value = 'Email is required.'
    return
  }

  busy.value = true
  try {
    await usersApi.forgotPassword({ email: e })
    statusMessage.value = 'Reset code sent. Check your email.'
    await router.push({ name: 'reset-password', query: { email: e } })
  } catch (err) {
    errorMessage.value = authErrorMessage(err)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <main class="page">
    <header class="header">
      <h1 class="title">Forgot password</h1>
      <p class="subtitle">We’ll email you a reset code.</p>
    </header>

    <section class="card">
      <h2 class="sectionTitle">Send reset code</h2>

      <form class="form" @submit.prevent="submit">
        <label class="field">
          <span class="label">Email</span>
          <input v-model="email" class="input" type="email" autocomplete="email" />
        </label>

        <div class="actions">
          <AppButton type="submit" :disabled="busy">{{ busy ? 'Sending…' : 'Send code' }}</AppButton>
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
