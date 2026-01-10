export type QrCode = {
  id: string
  label: string
  url: string
  active: boolean
  createdAtIso: string
  qrDataUrl?: string
}

export type CreateQrCodeInput = {
  label: string
  url: string
  active?: boolean
}

export type UpdateQrCodeInput = {
  label?: string
  url?: string
  active?: boolean
}
