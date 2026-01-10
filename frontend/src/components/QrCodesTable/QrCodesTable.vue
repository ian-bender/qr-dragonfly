<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue'
import { requestJson } from '../../api'
import type { QrCodeItem } from '../../types/qrCodeItem'
import { formatDateTime } from '../../lib/date'
import { generateQrDataUrl } from '../../lib/qr'

type DailyClicks = {
  qrCodeId: string
  dayIso: string
  total: number
  hour00: number
  hour01: number
  hour02: number
  hour03: number
  hour04: number
  hour05: number
  hour06: number
  hour07: number
  hour08: number
  hour09: number
  hour10: number
  hour11: number
  hour12: number
  hour13: number
  hour14: number
  hour15: number
  hour16: number
  hour17: number
  hour18: number
  hour19: number
  hour20: number
  hour21: number
  hour22: number
  hour23: number
}

type Props = {
  qrCodes: QrCodeItem[]
  updatingId: string | null
  errorMessage: string | null
  showSampleWhenEmpty?: boolean
}

const props = defineProps<Props>()

const isShowingSamples = computed(() => props.qrCodes.length === 0 && Boolean(props.showSampleWhenEmpty))

const sampleQrCodes: QrCodeItem[] = [
  {
    id: 'sample-1',
    label: 'Example menu',
    url: 'https://example.com/menu',
    active: true,
    createdAtIso: new Date().toISOString(),
    qrDataUrl: '',
  },
  {
    id: 'sample-2',
    label: 'Product page',
    url: 'https://example.com/product?utm_source=qr',
    active: true,
    createdAtIso: new Date().toISOString(),
    qrDataUrl: '',
  },
  {
    id: 'sample-3',
    label: 'Support link',
    url: 'https://example.com/support',
    active: false,
    createdAtIso: new Date().toISOString(),
    qrDataUrl: '',
  },
]

const rows = computed(() => {
  if (props.qrCodes.length > 0) return props.qrCodes
  if (isShowingSamples.value) return sampleQrCodes
  return []
})

const searchText = ref('')

const filteredRows = computed(() => {
  const q = searchText.value.trim().toLowerCase()
  if (!q) return rows.value
  return rows.value.filter((r) => {
    const label = (r.label ?? '').toLowerCase()
    const url = (r.url ?? '').toLowerCase()
    return label.includes(q) || url.includes(q)
  })
})

const PAGE_SIZE_KEY = 'qrCodesTable.pageSize'
const pageSize = ref<number>(10)
const page = ref<number>(1)

function loadPageSize(): number {
  try {
    const raw = window.localStorage.getItem(PAGE_SIZE_KEY)
    const n = raw ? Number(raw) : 10
    return n === 10 || n === 25 || n === 50 ? n : 10
  } catch {
    return 10
  }
}

function savePageSize(n: number) {
  try {
    window.localStorage.setItem(PAGE_SIZE_KEY, String(n))
  } catch {
    // ignore
  }
}

watchEffect(() => {
  // initialize once in the browser
  if (pageSize.value !== 10) return
  pageSize.value = loadPageSize()
})

watchEffect(() => {
  savePageSize(pageSize.value)
})

const totalRows = computed(() => filteredRows.value.length)
const totalPages = computed(() => Math.max(1, Math.ceil(totalRows.value / pageSize.value)))

watchEffect(() => {
  if (page.value > totalPages.value) page.value = totalPages.value
  if (page.value < 1) page.value = 1
})

const pageStartIndex = computed(() => (page.value - 1) * pageSize.value)
const pageEndIndexExclusive = computed(() => Math.min(totalRows.value, pageStartIndex.value + pageSize.value))
const pagedRows = computed(() => filteredRows.value.slice(pageStartIndex.value, pageEndIndexExclusive.value))

const canPrev = computed(() => page.value > 1)
const canNext = computed(() => page.value < totalPages.value)

function prevPage() {
  if (!canPrev.value) return
  page.value -= 1
}

function nextPage() {
  if (!canNext.value) return
  page.value += 1
}

watchEffect(() => {
  // Reset to page 1 on search changes.
  void searchText.value
  page.value = 1
})

const showError = computed(() => Boolean(props.errorMessage) && !isShowingSamples.value)

function isSampleId(id: string): boolean {
  return id.startsWith('sample-')
}

const sampleTrendPathById: Record<string, string> = {
  'sample-1': sparklinePath([1, 2, 3, 4, 6, 8, 11]),
  'sample-2': sparklinePath([0, 1, 1, 2, 3, 5, 7]),
  'sample-3': sparklinePath([2, 2, 3, 3, 4, 5, 6]),
}

function sampleTrendPath(id: string): string {
  return sampleTrendPathById[id] ?? sparklinePath([1, 2, 3, 4, 5, 6, 7])
}

const sampleQrDataUrlById = ref<Record<string, string>>({})

function qrImageSrc(qrCode: QrCodeItem): string {
  if (qrCode.qrDataUrl) return qrCode.qrDataUrl
  if (isSampleId(qrCode.id)) return sampleQrDataUrlById.value[qrCode.id] ?? ''
  return ''
}

function dateIsoUTC(daysAgo: number): string {
  const d = new Date()
  d.setUTCDate(d.getUTCDate() - daysAgo)
  return d.toISOString().slice(0, 10)
}

const todayIso = dateIsoUTC(0)

const last7DaysIso = computed(() => {
  // Oldest -> newest for left-to-right charts.
  return [dateIsoUTC(6), dateIsoUTC(5), dateIsoUTC(4), dateIsoUTC(3), dateIsoUTC(2), dateIsoUTC(1), dateIsoUTC(0)]
})

type TrendState = {
  dayIso: string
  counts: number[]
  total: number
  delta: number
  isUp: boolean
  pathD: string
}

const trendById = ref<Record<string, TrendState>>({})
const trendLoading = ref<Record<string, boolean>>({})

async function fetchDailyClicks(qrId: string, dayIso: string): Promise<DailyClicks | null> {
  try {
    return await requestJson<DailyClicks>({
      method: 'GET',
      path: `/api/clicks/${encodeURIComponent(qrId)}/daily`,
      query: { day: dayIso },
    })
  } catch {
    return null
  }
}

function sparklinePath(counts: number[], width = 96, height = 24, pad = 2): string {
  const n = counts.length
  if (n === 0) return ''

  const first = counts[0] ?? 0
  let min = first
  let max = first
  for (const v of counts) {
    if (v < min) min = v
    if (v > max) max = v
  }

  const usableW = Math.max(1, width - pad * 2)
  const usableH = Math.max(1, height - pad * 2)
  const dx = n === 1 ? 0 : usableW / (n - 1)

  const yFor = (v: number) => {
    if (max === min) return pad + usableH / 2
    const t = (v - min) / (max - min)
    return pad + (1 - t) * usableH
  }

  let d = ''
  for (let i = 0; i < n; i++) {
    const v = counts[i] ?? 0
    const x = pad + i * dx
    const y = yFor(v)
    d += i === 0 ? `M ${x.toFixed(2)} ${y.toFixed(2)}` : ` L ${x.toFixed(2)} ${y.toFixed(2)}`
  }
  return d
}

function trendTotal(qrId: string): number {
  return trendById.value[qrId]?.total ?? 0
}

function trendPath(qrId: string): string {
  return trendById.value[qrId]?.pathD ?? ''
}

function trendIsUp(qrId: string): boolean {
  return trendById.value[qrId]?.isUp ?? true
}

async function loadTrend(qrId: string) {
  if (trendLoading.value[qrId]) return
  if (trendById.value[qrId]?.dayIso === todayIso) return
  trendLoading.value = { ...trendLoading.value, [qrId]: true }

  try {
    const days = last7DaysIso.value
    const results = await Promise.all(days.map((d) => fetchDailyClicks(qrId, d)))
    const counts = results.map((r) => r?.total ?? 0)
    const total = counts.reduce((a, b) => a + b, 0)
    const delta = counts.length >= 2 ? (counts[counts.length - 1] ?? 0) - (counts[counts.length - 2] ?? 0) : 0
    const isUp = delta >= 0

    trendById.value = {
      ...trendById.value,
      [qrId]: {
        dayIso: todayIso,
        counts,
        total,
        delta,
        isUp,
        pathD: sparklinePath(counts),
      },
    }
  } finally {
    trendLoading.value = { ...trendLoading.value, [qrId]: false }
  }
}

watchEffect(() => {
  for (const row of pagedRows.value) {
    if (isSampleId(row.id)) continue
    void loadTrend(row.id)
  }
})

watchEffect(() => {
  if (!isShowingSamples.value) return
  for (const row of sampleQrCodes) {
    if (sampleQrDataUrlById.value[row.id]) continue
    void (async () => {
      try {
        const dataUrl = await generateQrDataUrl(row.url)
        sampleQrDataUrlById.value = { ...sampleQrDataUrlById.value, [row.id]: dataUrl }
      } catch {
        // ignore
      }
    })()
  }
})

const emit = defineEmits<{
  (e: 'copy-url', url: string): void
  (e: 'download', qrCode: QrCodeItem): void
  (e: 'remove', id: string): void
  (e: 'update', id: string, input: { label: string; url: string }): void
  (e: 'set-active', id: string, active: boolean): void
}>()

const editingId = ref<string | null>(null)
const editLabel = ref('')
const editUrl = ref('')

function startEdit(qrCode: QrCodeItem) {
  if (isSampleId(qrCode.id)) return
  editingId.value = qrCode.id
  editLabel.value = qrCode.label
  editUrl.value = qrCode.url
}

function cancelEdit() {
  editingId.value = null
  editLabel.value = ''
  editUrl.value = ''
}

function saveEdit() {
  const id = editingId.value
  if (!id) return
  if (isSampleId(id)) return
  emit('update', id, { label: editLabel.value, url: editUrl.value })
}

function isEditing(tabId: string): boolean {
  return editingId.value === tabId
}

function isBusy(tabId: string): boolean {
  return props.updatingId === tabId
}
</script>

<template>
  <section class="card">
    <div class="tableHeader">
      <h2 class="sectionTitle">QR codes</h2>
      <div class="meta">{{ totalRows }} total</div>
    </div>

    <p v-if="showError" class="error">{{ errorMessage }}</p>

    <div v-if="rows.length === 0" class="empty">No QR codes yet. Create one above.</div>

    <div v-else class="tableWrap">
      <div v-if="isShowingSamples" class="sampleNote">
        Sample rows shown while signed out.
      </div>

      <div class="pager">
        <div class="pagerLeft">
          <span class="pagerText">Showing {{ pageStartIndex + 1 }}–{{ pageEndIndexExclusive }} of {{ totalRows }}</span>
        </div>
        <div class="pagerRight">
          <input
            v-model="searchText"
            class="pagerSearch"
            type="search"
            placeholder="Search label or URL"
            aria-label="Search QR codes"
          />

          <label class="pagerText">
            Per page
            <select v-model.number="pageSize" class="pagerSelect" aria-label="Rows per page">
              <option :value="10">10</option>
              <option :value="25">25</option>
              <option :value="50">50</option>
            </select>
          </label>

          <button class="buttonSmall" type="button" :disabled="!canPrev" @click="prevPage">Prev</button>
          <span class="pagerText">Page {{ page }} / {{ totalPages }}</span>
          <button class="buttonSmall" type="button" :disabled="!canNext" @click="nextPage">Next</button>
        </div>
      </div>

      <table class="table">
        <thead>
          <tr>
            <th>Label</th>
            <th>URL</th>
            <th>Active</th>
            <th>Created</th>
            <th>Trend</th>
            <th>QR</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="qrCode in pagedRows" :key="qrCode.id">
            <td class="mono">
              <template v-if="isEditing(qrCode.id)">
                <input v-model="editLabel" class="input" type="text" />
              </template>
              <template v-else>
                {{ qrCode.label }}
              </template>
            </td>
            <td class="urlCell">
              <template v-if="isEditing(qrCode.id)">
                <input v-model="editUrl" class="input" type="url" inputmode="url" autocomplete="off" />
              </template>
              <template v-else>
                <span v-if="isSampleId(qrCode.id)" class="link sampleUrl" aria-label="Example URL">{{ qrCode.url }}</span>
                <a v-else class="link" :href="qrCode.url" target="_blank" rel="noreferrer">{{ qrCode.url }}</a>
              </template>
            </td>
            <td>
              <template v-if="isSampleId(qrCode.id)">
                <span class="muted">{{ qrCode.active ? 'Active' : 'Inactive' }}</span>
              </template>
              <button
                v-else
                class="buttonSmall"
                type="button"
                :disabled="isBusy(qrCode.id)"
                @click="emit('set-active', qrCode.id, !qrCode.active)"
              >
                {{ qrCode.active ? 'Active' : 'Inactive' }}
              </button>
            </td>
            <td class="muted">{{ formatDateTime(qrCode.createdAtIso) }}</td>
            <td>
              <div v-if="isSampleId(qrCode.id)" class="sparkWrap" aria-hidden="true">
                <svg class="spark" viewBox="0 0 100 28">
                  <path :d="sampleTrendPath(qrCode.id)" class="sparkLine up" fill="none" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" />
                </svg>
              </div>
              <div v-else class="sparkWrap" :title="`Clicks last 7 days: ${trendTotal(qrCode.id)}`">
                <svg class="spark" viewBox="0 0 100 28" role="img" aria-label="Click trend (last 7 days)">
                  <path
                    v-if="trendPath(qrCode.id)"
                    :d="trendPath(qrCode.id)"
                    :class="trendIsUp(qrCode.id) ? 'sparkLine up' : 'sparkLine down'"
                    fill="none"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                  />
                  <path v-else d="M 2 14 L 98 14" class="sparkLine flat" fill="none" stroke-linecap="round" stroke-width="2" />
                </svg>
              </div>
            </td>
            <td>
              <img v-if="qrImageSrc(qrCode)" class="qr" :src="qrImageSrc(qrCode)" :alt="`QR for ${qrCode.url}`" />
              <div v-else class="qrPlaceholder" aria-hidden="true">QR</div>
            </td>
            <td class="actions">
              <template v-if="isSampleId(qrCode.id)">
                <span class="muted">Sign in to manage</span>
              </template>
              <template v-else-if="isEditing(qrCode.id)">
                <button class="buttonSmall" type="button" :disabled="isBusy(qrCode.id)" @click="saveEdit">
                  {{ isBusy(qrCode.id) ? 'Saving…' : 'Save' }}
                </button>
                <button class="buttonSmall" type="button" :disabled="isBusy(qrCode.id)" @click="cancelEdit">
                  Cancel
                </button>
              </template>
              <template v-else>
                <button class="buttonSmall" type="button" :disabled="isBusy(qrCode.id)" @click="startEdit(qrCode)">
                  Edit
                </button>
                <button class="buttonSmall" type="button" :disabled="isBusy(qrCode.id)" @click="emit('copy-url', qrCode.url)">
                  Copy URL
                </button>
                <button class="buttonSmall" type="button" :disabled="isBusy(qrCode.id)" @click="emit('download', qrCode)">
                  Download
                </button>
                <button class="buttonSmall danger" type="button" :disabled="isBusy(qrCode.id)" @click="emit('remove', qrCode.id)">
                  Remove
                </button>
              </template>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="rows.length > 0 && totalRows === 0" class="empty" style="margin-top: 12px">
        No matches. Try a different search.
      </div>
    </div>
  </section>
</template>

<style scoped src="./QrCodesTable.scss" lang="scss"></style>
