import { computed, onMounted, ref } from 'vue'
import { qrCodesApi } from '../api'
import { useUser } from './useUser'
import { generateQrDataUrl } from '../lib/qr'
import { trackingUrlForQrId } from '../lib/tracking'
import type { QrCodeItem } from '../types/qrCodeItem'

function isValidUrl(value: string): boolean {
  try {
    const parsed = new URL(value)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

export function useQrCodes() {
  const qrCodes = ref<QrCodeItem[]>([])

  const labelInput = ref('')
  const urlInput = ref('')
  const isCreating = ref(false)
  const errorMessage = ref<string | null>(null)
  const isLoading = ref(false)
  const updatingId = ref<string | null>(null)

  const hasQrCodes = computed(() => qrCodes.value.length > 0)

  const { userType } = useUser()

  async function hydrateQrDataUrls(items: { id: string; url: string }[]): Promise<Record<string, string>> {
    const out: Record<string, string> = {}
    await Promise.all(
      items.map(async (i) => {
        try {
          out[i.id] = await generateQrDataUrl(trackingUrlForQrId(i.id))
        } catch {
          out[i.id] = ''
        }
      }),
    )
    return out
  }

  async function loadQrCodes(): Promise<void> {
    errorMessage.value = null
    isLoading.value = true
    try {
      const items = await qrCodesApi.list()
      const qrById = await hydrateQrDataUrls(items.map((i) => ({ id: i.id, url: i.url })))
      qrCodes.value = items.map((i) => ({
        id: i.id,
        label: i.label,
        url: i.url,
        active: i.active,
        createdAtIso: i.createdAtIso,
        qrDataUrl: qrById[i.id] || '',
      }))
    } catch {
      errorMessage.value = 'Failed to load QR codes from the server.'
    } finally {
      isLoading.value = false
    }
  }

  onMounted(() => {
    void loadQrCodes()
  })

  async function createQrCode(): Promise<void> {
    errorMessage.value = null

    const label = labelInput.value.trim() || 'Untitled'
    const url = urlInput.value.trim()

    if (!url) {
      errorMessage.value = 'Enter a URL.'
      return
    }

    if (!isValidUrl(url)) {
      errorMessage.value = 'Enter a valid http(s) URL (include https://).'
      return
    }

    isCreating.value = true
    try {
      const created = await qrCodesApi.create({ label, url, active: true }, userType.value)
      const qrDataUrl = await generateQrDataUrl(trackingUrlForQrId(created.id))
      const item: QrCodeItem = {
        id: created.id,
        label: created.label,
        url: created.url,
        active: created.active,
        createdAtIso: created.createdAtIso,
        qrDataUrl,
      }

      qrCodes.value = [item, ...qrCodes.value.filter((q) => q.id !== item.id)]
      labelInput.value = ''
      urlInput.value = ''
    } catch {
      errorMessage.value = 'Failed to create QR code.'
    } finally {
      isCreating.value = false
    }
  }

  async function deleteQrCode(id: string): Promise<void> {
    errorMessage.value = null
    try {
      await qrCodesApi.delete(id, userType.value)
      qrCodes.value = qrCodes.value.filter((q) => q.id !== id)
    } catch {
      errorMessage.value = 'Failed to delete QR code.'
    }
  }

  async function updateQrCode(id: string, input: { label: string; url: string }): Promise<void> {
    errorMessage.value = null

    const label = input.label.trim()
    const url = input.url.trim()

    if (!url) {
      errorMessage.value = 'Enter a URL.'
      return
    }
    if (!isValidUrl(url)) {
      errorMessage.value = 'Enter a valid http(s) URL (include https://).'
      return
    }

    const current = qrCodes.value.find((q) => q.id === id)
    if (!current) return

    const patch: { label?: string; url?: string } = {}
    if (label !== current.label) patch.label = label
    if (url !== current.url) patch.url = url

    // No-op
    if (!patch.label && !patch.url) return

    updatingId.value = id
    try {
      const updated = await qrCodesApi.update(id, patch, userType.value)
      const nextQrDataUrl = current.qrDataUrl

      qrCodes.value = qrCodes.value.map((q) =>
        q.id === id
          ? {
              ...q,
              label: updated.label,
              url: updated.url,
              active: updated.active,
              createdAtIso: updated.createdAtIso,
              qrDataUrl: nextQrDataUrl,
            }
          : q,
      )
    } catch {
      errorMessage.value = 'Failed to update QR code.'
    } finally {
      updatingId.value = null
    }
  }

  async function setQrCodeActive(id: string, active: boolean): Promise<void> {
    errorMessage.value = null

    const current = qrCodes.value.find((q) => q.id === id)
    if (!current) return
    if (current.active === active) return

    updatingId.value = id
    try {
      const updated = await qrCodesApi.update(id, { active }, userType.value)
      qrCodes.value = qrCodes.value.map((q) => (q.id === id ? { ...q, active: updated.active } : q))
    } catch {
      errorMessage.value = 'Failed to update QR code.'
    } finally {
      updatingId.value = null
    }
  }

  async function copyToClipboard(text: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(text)
    } catch {
      // ignore
    }
  }

  function downloadQrCode(qrCode: QrCodeItem): void {
    const link = document.createElement('a')
    link.href = qrCode.qrDataUrl
    const safeLabel = qrCode.label.trim().replace(/[^a-z0-9_-]+/gi, '_')
    link.download = `${safeLabel || 'qr'}_${qrCode.id}.png`
    link.click()
  }

  return {
    qrCodes,
    hasQrCodes,
    labelInput,
    urlInput,
    isCreating,
    isLoading,
    updatingId,
    errorMessage,
    loadQrCodes,
    createQrCode,
    updateQrCode,
    setQrCodeActive,
    deleteQrCode,
    copyToClipboard,
    downloadQrCode,
  }
}
