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

const busy = ref<null | 'confirm' | 'resend'>(null)
const errorMessage = ref<string | null>(null)
const statusMessage = ref<string | null>(null)

watchEffect(() => {
  const q = route.query.email
  if (typeof q === 'string' && !email.value) email.value = q
})

async function confirm() {
  errorMessage.value = null
  statusMessage.value = null

  const e = email.value.trim().toLowerCase()
  const c = code.value.trim()
  if (!e || !c) {
    errorMessage.value = 'Email and confirmation code are required.'
    return
  }

  busy.value = 'confirm'
  try {
    await usersApi.confirmSignUp({ email: e, code: c })
    statusMessage.value = 'Email confirmed. You can now log in.'
    await router.push({ name: 'login', query: { email: e } })
  } catch (err) {
    errorMessage.value = authErrorMessage(err)
  } finally {
    busy.value = null
  }
}

async function resend() {
  errorMessage.value = null
  statusMessage.value = null

  const e = email.value.trim().toLowerCase()
  if (!e) {
    errorMessage.value = 'Email is required.'
    return
  }

  busy.value = 'resend'
  try {
    await usersApi.resendConfirmation({ email: e })
    statusMessage.value = 'Confirmation code sent. Check your email.'
  } catch (err) {
    errorMessage.value = authErrorMessage(err)
  } finally {
    busy.value = null
  }
}
</script>

<template>
  <main class="page">
    <header class="header">
      <h1 class="title">Confirm email</h1>
      <p class="subtitle">Enter the confirmation code sent to your email.</p>
    </header>

    <section class="card">
      <h2 class="sectionTitle">Confirm</h2>

      <form class="form" @submit.prevent="confirm">
        <label class="field">
          <span class="label">Email</span>
          <input v-model="email" class="input" type="email" autocomplete="email" />
        </label>

        <label class="field">
          <span class="label">Confirmation code</span>
          <input v-model="code" class="input" type="text" inputmode="numeric" autocomplete="one-time-code" />
        </label>

        <div class="actions">
          <AppButton type="submit" :disabled="busy !== null">{{ busy === 'confirm' ? 'Confirming…' : 'Confirm' }}</AppButton>
          <AppButton type="button" :disabled="busy !== null" @click="resend">
            {{ busy === 'resend' ? 'Sending…' : 'Resend code' }}
          </AppButton>
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
