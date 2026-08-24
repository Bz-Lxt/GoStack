<script setup lang="ts">
import { useTelemetry } from '../stores/telemetry'
const store = useTelemetry()
</script>

<template>
  <aside class="panel p-4 h-full">
    <h2 class="font-display text-phosphor tracking-widest text-sm mb-3">连接</h2>
    <ul v-if="store.conns.length" class="space-y-2">
      <li v-for="c in store.conns" :key="c.id">
        <button
          type="button"
          class="w-full text-left px-3 py-2 rounded-lg border border-line hover:border-phosphor/40"
          :class="store.activeId === c.id ? 'bg-phosphor/10 border-phosphor/50' : ''"
          @click="store.activeId = c.id"
        >
          <div class="text-[11px] text-ice font-display">{{ c.state }}</div>
          <div class="text-[10px] break-all">{{ c.remote }} → {{ c.local }}</div>
          <div class="text-[10px] text-mist">rx {{ c.rx_bytes }} · tx {{ c.tx_bytes }} · rto {{ c.rto_ms }}ms</div>
        </button>
      </li>
    </ul>
    <p v-else class="text-xs text-mist/70">空态：还没有 TCB。点顶栏「放行流量」让内核或对端栈连入 10.0.0.2:9000。</p>
  </aside>
</template>
