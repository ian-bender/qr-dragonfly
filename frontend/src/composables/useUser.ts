import { computed, onMounted, ref } from 'vue'
import { usersApi } from '../api'
import type { User } from '../api'

const currentUser = ref<User | null>(null)
const isLoading = ref(false)
const isLoaded = ref(false)
const errorMessage = ref<string | null>(null)

async function loadCurrentUserOnce(): Promise<void> {
  if (isLoaded.value || isLoading.value) return
  isLoading.value = true
  errorMessage.value = null
  try {
    // If not authenticated, backend returns 401; treat as anonymous.
    const user = await usersApi.me().catch(() => null)
    currentUser.value = user
  } finally {
    isLoaded.value = true
    isLoading.value = false
  }
}

async function reloadCurrentUser(): Promise<void> {
  isLoaded.value = false
  await loadCurrentUserOnce()
}

export function useUser() {
  onMounted(() => {
    void loadCurrentUserOnce()
  })

  const userType = computed<"free" | "basic" | "enterprise" | "admin">(() => {
    const t = currentUser.value?.userType?.toLowerCase()
    if (t === 'basic' || t === 'enterprise' || t === 'admin') return t
    return 'free'
  })

  return {
    user: currentUser,
    isLoading,
    isLoaded,
    errorMessage,
    userType,
    reload: reloadCurrentUser,
  }
}
