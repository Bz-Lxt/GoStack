<script setup lang="ts">
import { computed } from 'vue'
import { useTelemetry } from '../stores/telemetry'

const store = useTelemetry()
const cells = computed(() => {
  if (store.cells.length) return store.cells
  return Array.from({ length: 32 }, (_, i) => ({ seq: i * 1460, len: 1460, mark: 'outside' }))
})

const color: Record<string, string> = {
  acked: 'bg-phosphor/80',
  inflight: 'bg-amber/80',
  usable: 'bg-ice/70',
  outside: 'bg-line',
  retrans: 'bg-danger animate-pulse',
}
</script>

<template>
  <section class="panel p-4">
    <header class="flex items-center justify-between mb-3">
      <h2 class="font-display text-phosphor tracking-widest text-sm">滑窗网格</h2>
      <div class="flex gap-3 text-[10px] text-mist">
        <span><i class="inline-block w-2 h-2 bg-phosphor mr-1" />已确认</span>
        <span><i class="inline-block w-2 h-2 bg-amber mr-1" />在飞</span>
        <span><i class="inline-block w-2 h-2 bg-ice mr-1" />可发</span>
        <span><i class="inline-block w-2 h-2 bg-danger mr-1" />重传</span>
      </div>
    </header>
    <div class="grid grid-cols-8 gap-1.5">
      <button
        v-for="c in cells"
        :key="c.seq"
        class="h-9 rounded-sm border border-white/5 text-[9px] font-mono text-ink/90"
        :class="color[c.mark] || color.outside"
        :title="`SEQ ${c.seq} · ${c.mark}`"
        type="button"
      >
        {{ c.seq % 100000 }}
      </button>
    </div>
    <p v-if="!store.cells.length" class="text-xs text-mist/60 mt-3">空态：尚无窗口快照。发起一次流量后格子会随 SND.UNA / NXT 滑动。</p>
  </section>
</template>
