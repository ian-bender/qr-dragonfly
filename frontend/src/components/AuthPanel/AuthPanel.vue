<script setup lang="ts">
import { computed, ref } from 'vue'
import { ApiError, usersApi } from '../../api'
import { useUser } from '../../composables/useUser'
import AppButton from '../ui/AppButton.vue'

function toErrorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    const payload = err.payload as any
    const code = payload?.error
    if (typeof code === 'string' && code.trim()) {
      switch (code) {
        case 'user_not_confirmed':
          return 'Email not confirmed yet. Check your inbox for the confirmation code, then confirm.'
        case 'code_mismatch':
          return 'That confirmation code is incorrect.'
        case 'code_expired':
          return 'That confirmation code has expired. Resend and try again.'
        case 'user_already_exists':
          return 'An account with that email already exists.'
        case 'not_authorized':
          return 'Invalid credentials, or the account is not allowed to sign in yet.'
        case 'rate_limited':
          return 'Too many attempts. Please wait and try again.'
        case 'invalid_password':
          return 'Password does not meet the requirements.'
        default:
          return code
      }
    }
    return `${err.status} ${err.message}`
  }
  if (err instanceof Error) return err.message
  return 'Request failed.'
}

const { user, isLoading: userLoading, reload } = useUser()

const isAuthed = computed(() => Boolean(user.value?.email))

const statusMessage = ref<string | null>(null)
const errorMessage = ref<string | null>(null)
const busy = ref<null | 'register' | 'confirm' | 'resend' | 'login' | 'logout' | 'forgot' | 'reset' | 'change'>(null)

// Register
const registerEmail = ref('')
const registerPassword = ref('')

// Confirm signup
const confirmEmail = ref('')
const confirmCode = ref('')

// Login
const loginEmail = ref('')
const loginPassword = ref('')

// Forgot/reset
const forgotEmail = ref('')
const resetEmail = ref('')
const resetCode = ref('')
const resetNewPassword = ref('')

// Change password
const oldPassword = ref('')
const newPassword = ref('')

async function run(action: typeof busy.value, fn: () => Promise<void>) {
  statusMessage.value = null
  errorMessage.value = null
  busy.value = action
  try {
    await fn()
  } catch (err) {
    errorMessage.value = toErrorMessage(err)
  } finally {
    busy.value = null
  }
}

async function onRegister() {
  await run('register', async () => {
    const email = registerEmail.value.trim().toLowerCase()
    const password = registerPassword.value.trim()

    if (!email || !password) {
      errorMessage.value = 'email_and_password_required'
      return
    }

    await usersApi.register({ email, password })
    statusMessage.value = 'Registered. Check your email for a confirmation code (if required), then confirm and log in.'

    confirmEmail.value = email
    forgotEmail.value = email
    resetEmail.value = email
    loginEmail.value = email

    registerPassword.value = ''
  })
}

async function onConfirm() {
  await run('confirm', async () => {
    const email = confirmEmail.value.trim().toLowerCase()
    const code = confirmCode.value.trim()
    if (!email || !code) {
      errorMessage.value = 'email_and_code_required'
      return
    }
    await usersApi.confirmSignUp({ email, code })
    statusMessage.value = 'Confirmed. You can now log in.'
  })
}

async function onResend() {
  await run('resend', async () => {
    const email = confirmEmail.value.trim().toLowerCase()
    if (!email) {
      errorMessage.value = 'email_required'
      return
    }
    const res = await usersApi.resendConfirmation({ email })
    const dest = res?.delivery?.destination
    const medium = res?.delivery?.medium
    statusMessage.value = dest || medium ? `Confirmation code sent${dest ? ` to ${dest}` : ''}${medium ? ` (${medium})` : ''}.` : 'Confirmation code sent (if the account exists and requires confirmation).'
  })
}

async function onLogin() {
  await run('login', async () => {
    const email = loginEmail.value.trim().toLowerCase()
    const password = loginPassword.value.trim()
    if (!email || !password) {
      errorMessage.value = 'email_and_password_required'
      return
    }
    await usersApi.login({ email, password })
    await reload()
    statusMessage.value = 'Logged in.'
    loginPassword.value = ''
  })
}

async function onLogout() {
  await run('logout', async () => {
    await usersApi.logout()
    await reload()
    statusMessage.value = 'Logged out.'
  })
}

async function onForgotPassword() {
  await run('forgot', async () => {
    const email = forgotEmail.value.trim().toLowerCase()
    if (!email) {
      errorMessage.value = 'email_required'
      return
    }
    await usersApi.forgotPassword({ email })
    statusMessage.value = 'Password reset code sent (if the account exists).'

    resetEmail.value = email
  })
}

async function onResetPassword() {
  await run('reset', async () => {
    const email = resetEmail.value.trim().toLowerCase()
    const code = resetCode.value.trim()
    const next = resetNewPassword.value.trim()
    if (!email || !code || !next) {
      errorMessage.value = 'email_code_new_password_required'
      return
    }
    await usersApi.confirmForgotPassword({ email, code, newPassword: next })
    statusMessage.value = 'Password reset. You can now log in with the new password.'

    loginEmail.value = email
    loginPassword.value = ''
    resetCode.value = ''
    resetNewPassword.value = ''
  })
}

async function onChangePassword() {
  await run('change', async () => {
    const oldPwd = oldPassword.value.trim()
    const newPwd = newPassword.value.trim()
    if (!oldPwd || !newPwd) {
      errorMessage.value = 'old_and_new_password_required'
      return
    }
    await usersApi.changePassword({ oldPassword: oldPwd, newPassword: newPwd })
    statusMessage.value = 'Password changed.'
    oldPassword.value = ''
    newPassword.value = ''
  })
}
</script>

<template>
  <section class="card">
    <div class="headerRow">
      <h2 class="sectionTitle">Account</h2>
      <div class="meta">
        <span v-if="userLoading">Loading…</span>
        <span v-else-if="isAuthed">Signed in as {{ user?.email }}</span>
        <span v-else>Not signed in</span>
      </div>
    </div>

    <p v-if="statusMessage" class="status">{{ statusMessage }}</p>
    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div class="grid">
      <form class="box" @submit.prevent="onRegister">
        <h3 class="boxTitle">Create account</h3>

        <label class="field">
          <span class="label">Email</span>
          <input v-model="registerEmail" class="input" type="email" autocomplete="email" />
        </label>

        <label class="field">
          <span class="label">Password</span>
          <input v-model="registerPassword" class="input" type="password" autocomplete="new-password" />
        </label>

        <AppButton type="submit" :disabled="busy !== null">{{ busy === 'register' ? 'Creating…' : 'Create account' }}</AppButton>
      </form>

      <form class="box" @submit.prevent="onConfirm">
        <h3 class="boxTitle">Confirm email</h3>

        <label class="field">
          <span class="label">Email</span>
          <input v-model="confirmEmail" class="input" type="email" autocomplete="email" />
        </label>

        <label class="field">
          <span class="label">Confirmation code</span>
          <input v-model="confirmCode" class="input" type="text" inputmode="numeric" autocomplete="one-time-code" />
        </label>

        <div class="row">
          <AppButton type="submit" :disabled="busy !== null">{{ busy === 'confirm' ? 'Confirming…' : 'Confirm' }}</AppButton>
          <AppButton type="button" :disabled="busy !== null" @click="onResend">{{ busy === 'resend' ? 'Sending…' : 'Resend code' }}</AppButton>
        </div>
      </form>

      <form class="box" @submit.prevent="onLogin">
        <h3 class="boxTitle">Login</h3>

        <label class="field">
          <span class="label">Email</span>
          <input v-model="loginEmail" class="input" type="email" autocomplete="email" />
        </label>

        <label class="field">
          <span class="label">Password</span>
          <input v-model="loginPassword" class="input" type="password" autocomplete="current-password" />
        </label>

        <div class="row">
          <AppButton type="submit" :disabled="busy !== null">{{ busy === 'login' ? 'Logging in…' : 'Login' }}</AppButton>
          <AppButton type="button" :disabled="busy !== null || !isAuthed" @click="onLogout">{{ busy === 'logout' ? 'Logging out…' : 'Logout' }}</AppButton>
        </div>
      </form>

      <form class="box" @submit.prevent="onChangePassword">
        <h3 class="boxTitle">Change password</h3>

        <label class="field">
          <span class="label">Old password</span>
          <input v-model="oldPassword" class="input" type="password" autocomplete="current-password" />
        </label>

        <label class="field">
          <span class="label">New password</span>
          <input v-model="newPassword" class="input" type="password" autocomplete="new-password" />
        </label>

        <AppButton type="submit" :disabled="busy !== null || !isAuthed">{{ busy === 'change' ? 'Updating…' : 'Change password' }}</AppButton>
      </form>

      <form class="box" @submit.prevent="onForgotPassword">
        <h3 class="boxTitle">Forgot password</h3>

        <label class="field">
          <span class="label">Email</span>
          <input v-model="forgotEmail" class="input" type="email" autocomplete="email" />
        </label>

        <AppButton type="submit" :disabled="busy !== null">{{ busy === 'forgot' ? 'Sending…' : 'Send reset code' }}</AppButton>
      </form>

      <form class="box" @submit.prevent="onResetPassword">
        <h3 class="boxTitle">Reset password</h3>

        <label class="field">
          <span class="label">Email</span>
          <input v-model="resetEmail" class="input" type="email" autocomplete="email" />
        </label>

        <label class="field">
          <span class="label">Reset code</span>
          <input v-model="resetCode" class="input" type="text" inputmode="numeric" autocomplete="one-time-code" />
        </label>

        <label class="field">
          <span class="label">New password</span>
          <input v-model="resetNewPassword" class="input" type="password" autocomplete="new-password" />
        </label>

        <AppButton type="submit" :disabled="busy !== null">{{ busy === 'reset' ? 'Resetting…' : 'Reset password' }}</AppButton>
      </form>
    </div>
  </section>
</template>

<style scoped src="./AuthPanel.scss" lang="scss"></style>
