<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { settingsApi, type UserSettings } from '../../api'
import { useUser } from '../../composables/useUser'

const { userType } = useUser()

const defaultRedirectUrl = ref('')
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
    console.log('[DefaultRedirectSettings] Fetching settings...')
    const settings = await settingsApi.get(userType.value)
    console.log('[DefaultRedirectSettings] Settings loaded:', settings)
    defaultRedirectUrl.value = settings.defaultRedirectUrl || ''
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
  console.log('[DefaultRedirectSettings] saveSettings called')
  if (!userType.value) {
    console.log('[DefaultRedirectSettings] No userType, cannot save')
    return
  }
  
  isSaving.value = true
  errorMessage.value = null
  successMessage.value = null
  
  try {
    const settings: UserSettings = {
      defaultRedirectUrl: defaultRedirectUrl.value.trim(),
    }
    console.log('[DefaultRedirectSettings] Saving settings:', settings)
    await settingsApi.update(settings, userType.value)
    console.log('[DefaultRedirectSettings] Settings saved successfully')
    successMessage.value = 'Settings saved successfully!'
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
      </form>

      <p v-if="successMessage" class="success">{{ successMessage }}</p>
      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </template>
  </section>
</template>

<style scoped lang="scss">
@use '../../styles/variables' as *;

.card {
  border: 1px solid $border-color;
  border-radius: $radius-lg;
  padding: $space-lg;
  margin-top: 14px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: $space-sm;
}

.sectionTitle {
  margin: 0;
  font-size: $font-size-base;
}

.expandButton {
  border-radius: $radius-md;
  padding: 6px 12px;
  border: 1px solid $border-color;
  background: rgba(0, 0, 0, 0.25);
  color: inherit;
  cursor: pointer;
  font-size: $font-size-xs;
  opacity: 0.8;
  transition: opacity 0.2s;

  &:hover {
    opacity: 1;
  }
}

.collapsedView {
  margin-top: $space-xs;
}

.currentUrl {
  margin: 0;
  padding: 8px 12px;
  background: rgba(0, 0, 0, 0.15);
  border-radius: $radius-md;
  font-size: $font-size-sm;
  opacity: 0.9;
  word-break: break-all;
}

.description {
  margin: 0 0 $space-lg;
  opacity: 0.8;
  font-size: $font-size-sm;
}

.form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: $space-md;
  align-items: end;
}

.field {
  display: grid;
  gap: $space-xs;
  min-width: 0;
}

.label {
  font-size: $font-size-xs;
  opacity: 0.8;
}

.input {
  width: 100%;
  box-sizing: border-box;
  border-radius: $radius-md;
  border: 1px solid $border-color;
  padding: 10px 12px;
  background: rgba(0, 0, 0, 0.15);
  color: inherit;
}

.input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.button {
  border-radius: $radius-md;
  padding: 10px 14px;
  border: 1px solid $border-color;
  background: rgba(0, 0, 0, 0.25);
  color: inherit;
  cursor: pointer;
}

.button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.success {
  margin: 10px 0 0;
  color: $color-success;
}

.error {
  margin: 10px 0 0;
  color: $color-error;
}

@media (max-width: 640px) {
  .form {
    grid-template-columns: 1fr;
  }
}
</style>
