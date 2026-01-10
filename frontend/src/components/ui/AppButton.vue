<script setup lang="ts">
import { computed, useAttrs } from 'vue'
import { RouterLink } from 'vue-router'

type Variant = 'primary' | 'secondary'

const props = withDefaults(
  defineProps<{
    to?: string | Record<string, any>
    type?: 'button' | 'submit' | 'reset'
    variant?: Variant
    disabled?: boolean
  }>(),
  {
    type: 'button',
    variant: 'primary',
    disabled: false,
  },
)

const attrs = useAttrs()

const forwardedAttrs = computed(() => {
  const { class: _class, ...rest } = attrs as Record<string, unknown>
  return rest
})

const className = computed(() => {
  const out: string[] = ['button']
  if (props.variant === 'secondary') out.push('secondary')
  if (typeof attrs.class === 'string' && attrs.class.trim()) out.push(attrs.class)
  return out.join(' ')
})
</script>

<template>
  <RouterLink v-if="to" :to="to" :class="className" v-bind="forwardedAttrs">
    <slot />
  </RouterLink>

  <button v-else :type="type" :disabled="disabled" :class="className" v-bind="forwardedAttrs">
    <slot />
  </button>
</template>
