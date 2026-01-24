<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { settingsApi, type UserSettings } from '../../api'
import { useUser } from '../../composables/useUser'

const { userType } = useUser()

const defaultRedirectUrl = ref('')
const originalUrl = ref('')
const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref<string | null>(null)
const successMessage = ref<string | null>(null)
const isCollapsed = ref(false)

console.log('[DefaultRedirectSettings] Component mounted, userType:', userType.value)

async function loadSettings() {
  console.log('[DefaultRedirectSettings] loadSettings called, userType:', userType.value)
  if (!userType.value) {
    console.log('[DefaultRedirectSettings] No userType, skipping load')
    return
  }
  
  isLoading.value = true
  errorMessage.value = null
  
  try {
    const settings = await settingsApi.get(userType.value)
    defaultRedirectUrl.value = settings.defaultRedirectUrl || ''
    originalUrl.value = settings.defaultRedirectUrl || ''
    // Auto-collapse if URL is set
    if (defaultRedirectUrl.value) {
      isCollapsed.value = true
    }
  } catch (err) {
    console.error('[DefaultRedirectSettings] Failed to load settings:', err)
    errorMessage.value = 'Failed to load settings'
  } finally {
    isLoading.value = false
  }
}

async function saveSettings() {
  if (!userType.value) {
    return
  }
  
  isSaving.value = true
  errorMessage.value = null
  successMessage.value = null
  
  try {
    const settings: UserSettings = {
      defaultRedirectUrl: defaultRedirectUrl.value.trim(),
    }
    await settingsApi.update(settings, userType.value)
    successMessage.value = 'Settings saved successfully!'
    originalUrl.value = defaultRedirectUrl.value.trim()
    // Auto-collapse after saving if URL is set
    if (defaultRedirectUrl.value.trim()) {
      isCollapsed.value = true
    }
    setTimeout(() => {
      successMessage.value = null
    }, 3000)
  } catch (err) {
    console.error('[DefaultRedirectSettings] Failed to save settings:', err)
    errorMessage.value = 'Failed to save settings'
  } finally {
    isSaving.value = false
  }
}

function cancelEdit() {
  defaultRedirectUrl.value = originalUrl.value
  errorMessage.value = null
  successMessage.value = null
  if (originalUrl.value) {
    isCollapsed.value = true
  }
}

onMounted(() => {
  console.log('[DefaultRedirectSettings] onMounted hook, calling loadSettings')
  loadSettings()
})

watch(userType, (newVal, oldVal) => {
  console.log('[DefaultRedirectSettings] userType changed from', oldVal, 'to', newVal)
  loadSettings()
})
</script>

<template>
  <section class="card">
    <div class="header">
      <h2 class="sectionTitle">Default Redirect URL</h2>
      <button 
        v-if="defaultRedirectUrl && isCollapsed" 
        class="expandButton"
        type="button"
        @click="isCollapsed = false"
      >
        Edit
      </button>
    </div>
    
    <div v-if="isCollapsed && defaultRedirectUrl" class="collapsedView">
      <p class="currentUrl">{{ defaultRedirectUrl }}</p>
    </div>

    <template v-else>
      <p class="description">
        When a QR code is inactive, users will be redirected to this URL instead of seeing a 404 error.
      </p>

      <form class="form" @submit.prevent="saveSettings">
        <label class="field">
          <span class="label">Default URL (optional)</span>
          <input
            v-model="defaultRedirectUrl"
            class="input"
            type="url"
            inputmode="url"
            autocomplete="off"
            placeholder="https://example.com/maintenance"
            :disabled="isLoading || isSaving"
          />
        </label>

        <button class="button" type="submit" :disabled="isLoading || isSaving">
          {{ isSaving ? 'Saving…' : 'Save Settings' }}
        </button>
        <button class="button secondary" type="button" :disabled="isLoading || isSaving" @click="cancelEdit">
          Cancel
        </button>
      </form>

      <p v-if="successMessage" class="success">{{ successMessage }}</p>
      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </template>
  </section>
</template>

<style scoped src="./DefaultRedirectSettings.scss" lang="scss"></style>
