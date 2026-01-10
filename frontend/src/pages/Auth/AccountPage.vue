<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { usersApi } from '../../api'
import { authErrorMessage } from '../../lib/authErrors'
import { useUser } from '../../composables/useUser'

const router = useRouter()
const { user, reload } = useUser()

const isAuthed = computed(() => Boolean(user.value?.email))

const busyLogout = ref(false)
const busyRefresh = ref(false)
const errorMessage = ref<string | null>(null)

const changePasswordDialog = ref<HTMLDialogElement | null>(null)
const oldPassword = ref('')
const newPassword = ref('')
const busyChangePassword = ref(false)
const changePasswordError = ref<string | null>(null)
const changePasswordStatus = ref<string | null>(null)

function openChangePassword() {
  changePasswordError.value = null
  changePasswordStatus.value = null
  oldPassword.value = ''
  newPassword.value = ''
  changePasswordDialog.value?.showModal()
}

function closeChangePassword() {
  changePasswordDialog.value?.close()
}

async function submitChangePassword() {
  changePasswordError.value = null
  changePasswordStatus.value = null

  if (!isAuthed.value) {
    await router.push({ name: 'login' })
    return
  }

  const oldPwd = oldPassword.value.trim()
  const newPwd = newPassword.value.trim()
  if (!oldPwd || !newPwd) {
    changePasswordError.value = 'Old and new passwords are required.'
    return
  }

  busyChangePassword.value = true
  try {
    await usersApi.changePassword({ oldPassword: oldPwd, newPassword: newPwd })
    await reload()
    changePasswordStatus.value = 'Password changed.'
    oldPassword.value = ''
    newPassword.value = ''
  } catch (err) {
    changePasswordError.value = authErrorMessage(err)
  } finally {
    busyChangePassword.value = false
  }
}

async function refreshAccount() {
  errorMessage.value = null
  busyRefresh.value = true
  try {
    await reload()
  } catch (err) {
    errorMessage.value = authErrorMessage(err)
  } finally {
    busyRefresh.value = false
  }
}

async function logout() {
  errorMessage.value = null
  busyLogout.value = true
  try {
    await usersApi.logout()
    await reload()
    await router.push({ name: 'home' })
  } catch (err) {
    errorMessage.value = authErrorMessage(err)
  } finally {
    busyLogout.value = false
  }
}
</script>

<template>
  <main class="page">
    <header class="header">
      <h1 class="title">Account</h1>
      <p class="subtitle">Manage your account settings.</p>
    </header>

    <section class="card">
      <h2 class="sectionTitle">Overview</h2>

      <p class="muted" v-if="!isAuthed">You are not signed in.</p>
      <div v-else class="kv">
        <div class="kvRow">
          <span class="kvKey">Email</span>
          <span class="kvVal">{{ user?.email }}</span>
        </div>
        <div class="kvRow" v-if="user?.id">
          <span class="kvKey">User ID</span>
          <span class="kvVal">{{ user?.id }}</span>
        </div>
      </div>

      <div class="links">
        <RouterLink v-if="!isAuthed" to="/login">Login</RouterLink>
        <RouterLink v-if="!isAuthed" to="/register">Create account</RouterLink>
        <RouterLink v-if="!isAuthed" to="/confirm">Confirm account</RouterLink>
        <RouterLink v-if="!isAuthed" to="/forgot-password">Forgot password</RouterLink>
        <RouterLink v-if="!isAuthed" to="/reset-password">Reset password</RouterLink>
      </div>

      <div v-if="isAuthed" class="divider" />

      <div v-if="isAuthed">
        <h2 class="sectionTitle" style="margin-top: 0">Security</h2>

        <div class="actions">
          <button class="button" type="button" @click="openChangePassword">Change password</button>
          <button class="button secondary" type="button" :disabled="busyRefresh" @click="refreshAccount">
            {{ busyRefresh ? 'Refreshing…' : 'Refresh' }}
          </button>
        </div>

        <div class="actions" style="margin-top: 12px">
          <button class="button" type="button" :disabled="busyLogout" @click="logout">
            {{ busyLogout ? 'Logging out…' : 'Logout' }}
          </button>
        </div>
      </div>

      <div v-if="!isAuthed" class="actions" style="margin-top: 12px">
        <RouterLink class="button" to="/login">Go to login</RouterLink>
      </div>

      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </section>

    <dialog ref="changePasswordDialog" class="dialog" @cancel.prevent="closeChangePassword">
      <div class="dialogInner">
        <div class="dialogHeader">
          <h2 class="dialogTitle">Change password</h2>
          <button class="iconButton" type="button" aria-label="Close" @click="closeChangePassword">×</button>
        </div>

        <form class="form" @submit.prevent="submitChangePassword">
          <label class="field">
            <span class="label">Old password</span>
            <input v-model="oldPassword" class="input" type="password" autocomplete="current-password" />
          </label>

          <label class="field">
            <span class="label">New password</span>
            <input v-model="newPassword" class="input" type="password" autocomplete="new-password" />
          </label>

          <div class="actions">
            <button class="button" type="submit" :disabled="busyChangePassword">
              {{ busyChangePassword ? 'Updating…' : 'Update password' }}
            </button>
            <button class="button secondary" type="button" :disabled="busyChangePassword" @click="closeChangePassword">
              Close
            </button>
          </div>
        </form>

        <p v-if="changePasswordStatus" class="status">{{ changePasswordStatus }}</p>
        <p v-if="changePasswordError" class="error">{{ changePasswordError }}</p>
      </div>
    </dialog>
  </main>
</template>

<style scoped src="./AuthPage.scss" lang="scss"></style>

<style scoped lang="scss">
.kv {
  display: grid;
  gap: 8px;
  margin-top: 12px;
}

.kvRow {
  display: flex;
  gap: 12px;
  align-items: baseline;
  flex-wrap: wrap;
}

.kvKey {
  opacity: 0.8;
  min-width: 72px;
}

.kvVal {
  font-weight: 600;
  word-break: break-word;
}

.divider {
  height: 1px;
  margin: 16px 0;
  background: color-mix(in srgb, $color-fg 12%, transparent);
}

.secondary {
  border: 1px solid color-mix(in srgb, $color-fg 12%, transparent);
  background: color-mix(in srgb, $color-bg 55%, transparent);
}

.dialog {
  width: min(560px, calc(100vw - 24px));
  border: 1px solid color-mix(in srgb, $color-fg 14%, transparent);
  border-radius: $radius-md;
  padding: 0;
  background: $color-bg;
  color: $color-fg;
}

.dialog::backdrop {
  background: rgba(0, 0, 0, 0.55);
}

.dialogInner {
  padding: 16px;
}

.dialogHeader {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.dialogTitle {
  margin: 0;
  font-size: 18px;
}

.iconButton {
  width: 36px;
  height: 36px;
  border-radius: $radius-md;
  border: 1px solid color-mix(in srgb, $color-fg 12%, transparent);
  background: $color-surface;
  color: $color-fg;
  padding: 0;
  font-size: 20px;
  line-height: 1;
}
</style>
