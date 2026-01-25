<script setup lang="ts">
import { computed } from 'vue'

type Props = {
  label: string
  url: string
  isCreating: boolean
  errorMessage: string | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'update:label', value: string): void
  (e: 'update:url', value: string): void
  (e: 'submit'): void
}>()

const labelModel = computed({
  get: () => props.label,
  set: (value: string) => emit('update:label', value),
})

const urlModel = computed({
  get: () => props.url,
  set: (value: string) => emit('update:url', value),
})
</script>

<template>
  <section class="card">
    <h2 class="sectionTitle">Create a QR code</h2>

    <form class="form" @submit.prevent="emit('submit')">
      <label class="field">
        <span class="label">Label</span>
        <input v-model="labelModel" class="input" type="text" placeholder="e.g. Landing page" />
      </label>

      <label class="field fieldWide">
        <span class="label">URL</span>
        <input
          v-model="urlModel"
          class="input"
          type="url"
          inputmode="url"
          autocomplete="off"
          placeholder="https://example.com"
        />
      </label>

      <div class="buttonWrapper">
        <button class="button" type="submit" :disabled="isCreating" tabindex="0">
          {{ isCreating ? 'Creating…' : 'Create' }}
        </button>
      </div>
    </form>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
  </section>
</template>

<style scoped src="./CreateQrCodeForm.scss" lang="scss"></style>
