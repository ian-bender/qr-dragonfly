const CLICK_BASE_URL: string = (import.meta as any).env?.VITE_CLICK_BASE_URL ?? 'https://qr-dragonfly.com'

export function trackingUrlForQrId(id: string): string {
  const base = CLICK_BASE_URL.replace(/\/+$/, '')
  return `${base}/r/${encodeURIComponent(id)}`
}
