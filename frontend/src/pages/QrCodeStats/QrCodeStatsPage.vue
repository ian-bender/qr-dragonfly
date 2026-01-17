<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { qrCodesApi, requestJson } from '../../api'
import { useUser } from '../../composables/useUser'
import { trackingUrlForQrId } from '../../lib/tracking'

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

function dateIsoUTC(daysAgo: number): string {
  const d = new Date()
  d.setUTCDate(d.getUTCDate() - daysAgo)
  return d.toISOString().slice(0, 10)
}

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

function hoursArray(dc: DailyClicks): number[] {
  return [
    dc.hour00,
    dc.hour01,
    dc.hour02,
    dc.hour03,
    dc.hour04,
    dc.hour05,
    dc.hour06,
    dc.hour07,
    dc.hour08,
    dc.hour09,
    dc.hour10,
    dc.hour11,
    dc.hour12,
    dc.hour13,
    dc.hour14,
    dc.hour15,
    dc.hour16,
    dc.hour17,
    dc.hour18,
    dc.hour19,
    dc.hour20,
    dc.hour21,
    dc.hour22,
    dc.hour23,
  ]
}

const router = useRouter()
const route = useRoute()

const id = computed(() => String(route.params.id ?? ''))

const { isAuthed, isLoaded, userType } = useUser()

watchEffect(() => {
  if (!isLoaded.value) return
  if (isAuthed.value) return
  const redirect = route.fullPath || '/'
  void router.replace({ name: 'login', query: { redirect } })
})

const qrCode = ref<{ id: string; label: string; url: string; active: boolean } | null>(null)
const isLoading = ref(false)
const errorMessage = ref<string | null>(null)

const last7Days = computed(() => [dateIsoUTC(6), dateIsoUTC(5), dateIsoUTC(4), dateIsoUTC(3), dateIsoUTC(2), dateIsoUTC(1), dateIsoUTC(0)])
const selectedDayIso = ref<string>(dateIsoUTC(0))

const dailyByDay = ref<Record<string, DailyClicks | null>>({})

const last7Total = computed(() => {
  return last7Days.value.reduce((sum, day) => sum + (dailyByDay.value[day]?.total ?? 0), 0)
})

const selectedDaily = computed(() => dailyByDay.value[selectedDayIso.value] ?? null)

const hourlyRows = computed(() => {
  const d = selectedDaily.value
  if (!d) return []
  return hoursArray(d).map((count, hour) => ({ hour, count }))
})

watchEffect(() => {
  if (!isAuthed.value) return
  if (!id.value) return

  errorMessage.value = null
  isLoading.value = true
  void (async () => {
    try {
      const item = await qrCodesApi.getById(id.value, userType.value)
      qrCode.value = { id: item.id, label: item.label, url: item.url, active: item.active }

      // Fetch daily click buckets for the last 7 days.
      const results = await Promise.all(last7Days.value.map((dayIso) => fetchDailyClicks(id.value, dayIso)))
      const next: Record<string, DailyClicks | null> = {}
      last7Days.value.forEach((dayIso, idx) => {
        next[dayIso] = results[idx] ?? null
      })
      dailyByDay.value = next

      // Ensure selected day is still in range.
      if (!last7Days.value.includes(selectedDayIso.value)) {
        selectedDayIso.value = last7Days.value[last7Days.value.length - 1] ?? dateIsoUTC(0)
      }
    } catch {
      errorMessage.value = 'Failed to load stats for this QR code.'
      qrCode.value = null
      dailyByDay.value = {}
    } finally {
      isLoading.value = false
    }
  })()
})
</script>

<template>
  <main class="page">
    <header class="header">
      <h1 class="title">QR Code Stats</h1>
      <p class="subtitle" v-if="qrCode">{{ qrCode.label }}</p>
      <p class="subtitle" v-else>Performance details for your QR code.</p>
    </header>

    <section class="card">
      <div class="topRow">
        <RouterLink class="link" to="/">← Back to QR codes</RouterLink>
        <div class="spacer" />
        <span v-if="qrCode" class="pill" :class="qrCode.active ? 'ok' : 'off'">{{ qrCode.active ? 'Active' : 'Inactive' }}</span>
      </div>

      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

      <div v-if="isLoading" class="muted">Loading…</div>

      <template v-else-if="qrCode">
        <div class="kv">
          <div class="kvRow">
            <span class="kvKey">Target</span>
            <a class="link" :href="qrCode.url" target="_blank" rel="noreferrer">{{ qrCode.url }}</a>
          </div>
          <div class="kvRow">
            <span class="kvKey">Tracking</span>
            <span class="mono">{{ trackingUrlForQrId(qrCode.id) }}</span>
          </div>
        </div>

        <div class="statsGrid">
          <div class="statCard">
            <div class="statLabel">Clicks (last 7 days)</div>
            <div class="statValue">{{ last7Total }}</div>
          </div>

          <div class="statCard">
            <div class="statLabel">Day</div>
            <select v-model="selectedDayIso" class="select" aria-label="Select day">
              <option v-for="d in last7Days" :key="d" :value="d">{{ d }}</option>
            </select>
          </div>

          <div class="statCard">
            <div class="statLabel">Clicks (selected day)</div>
            <div class="statValue">{{ selectedDaily?.total ?? 0 }}</div>
          </div>
        </div>

        <h2 class="sectionTitle">Hourly breakdown</h2>
        <p class="muted" v-if="!selectedDaily">No click data available for this day.</p>

        <table v-else class="table">
          <thead>
            <tr>
              <th>Hour (UTC)</th>
              <th>Clicks</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in hourlyRows" :key="row.hour">
              <td class="mono">{{ String(row.hour).padStart(2, '0') }}:00</td>
              <td class="mono">{{ row.count }}</td>
            </tr>
          </tbody>
        </table>
      </template>

      <template v-else>
        <p class="muted">No QR code found.</p>
      </template>
    </section>
  </main>
</template>

<style scoped src="../HomePage/HomePage.scss" lang="scss"></style>
<style scoped src="./QrCodeStatsPage.scss" lang="scss"></style>
