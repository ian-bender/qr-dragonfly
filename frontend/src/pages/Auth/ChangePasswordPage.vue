<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { usersApi } from '../../api'
import { authErrorMessage } from '../../lib/authErrors'
import { useUser } from '../../composables/useUser'
import AppButton from '../../components/ui/AppButton.vue'

const router = useRouter()
const { user, reload } = useUser()

const isAuthed = computed(() => Boolean(user.value?.email))

const oldPassword = ref('')
const newPassword = ref('')

const busy = ref(false)
const errorMessage = ref<string | null>(null)
const statusMessage = ref<string | null>(null)

async function submit() {
  errorMessage.value = null
  statusMessage.value = null

  if (!isAuthed.value) {
    await router.push({ name: 'login' })
    return
  }

  const oldPwd = oldPassword.value.trim()
  const newPwd = newPassword.value.trim()
  if (!oldPwd || !newPwd) {
    errorMessage.value = 'Old and new passwords are required.'
    return
  }

  busy.value = true
  try {
    await usersApi.changePassword({ oldPassword: oldPwd, newPassword: newPwd })
    await reload()
    statusMessage.value = 'Password changed.'
    oldPassword.value = ''
    newPassword.value = ''
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
      <h1 class="title">Change password</h1>
      <p class="subtitle">Update your password for this account.</p>
    </header>

    <section class="card">
      <h2 class="sectionTitle">Change password</h2>

      <p v-if="!isAuthed" class="muted">You must be logged in to change your password.</p>

      <form v-if="isAuthed" class="form" @submit.prevent="submit">
        <label class="field">
          <span class="label">Old password</span>
          <input v-model="oldPassword" class="input" type="password" autocomplete="current-password" />
        </label>

        <label class="field">
          <span class="label">New password</span>
          <input v-model="newPassword" class="input" type="password" autocomplete="new-password" />
        </label>

        <div class="actions">
          <AppButton type="submit" :disabled="busy">{{ busy ? 'Updating…' : 'Change password' }}</AppButton>
          <RouterLink class="muted" to="/account">Back to account</RouterLink>
        </div>
      </form>

      <p v-if="statusMessage" class="status">{{ statusMessage }}</p>
      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

      <div v-if="!isAuthed" class="links">
        <RouterLink to="/login">Login</RouterLink>
      </div>
    </section>
  </main>
</template>

<style scoped src="./AuthPage.scss" lang="scss"></style>
