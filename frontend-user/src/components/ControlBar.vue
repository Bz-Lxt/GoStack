<script setup lang="ts">
import { ref } from 'vue'
import { api } from '../lib/api'
import { useTelemetry } from '../stores/telemetry'

const store = useTelemetry()
const step = ref(false)
const pps = ref(10)
const loss = ref(0)
const busy = ref(false)

async function go() {
  busy.value = true
  try {
    await api.step(step.value, pps.value)
    await api.impair(loss.value, 0, 0, 0)
    const r = await api.traffic(32 * 1024)
    store.flash(`流量完成 · match=${(r as any).match}`)
    await store.refresh()
  } catch (e) {
    store.flash((e as Error).message, true)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="flex flex-wrap items-center gap-3">
    <label class="text-[11px] flex items-center gap-2">
      <input v-model="step" type="checkbox" class="accent-phosphor" />
      单步
    </label>
    <label class="text-[11px] flex items-center gap-2">
      pps
      <input v-model.number="pps" type="range" min="1" max="20" />
      <span class="w-6">{{ pps }}</span>
    </label>
    <label class="text-[11px] flex items-center gap-2">
      丢包
      <input v-model.number="loss" type="range" min="0" max="0.2" step="0.01" />
      <span>{{ (loss * 100).toFixed(0) }}%</span>
    </label>
    <button
      type="button"
      class="px-4 py-1.5 rounded-md bg-phosphor text-ink font-display text-sm disabled:opacity-40"
      :disabled="busy"
      @click="go"
    >
      {{ busy ? '发送中…' : '放行流量' }}
    </button>
  </div>
</template>
