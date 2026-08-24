<script setup lang="ts">
import { computed } from 'vue'
import { useTelemetry } from '../stores/telemetry'

const store = useTelemetry()
const hex = computed(() => String(store.selected?.payload.hex || ''))
const groups = computed(() => hex.value.match(/.{1,2}/g) || [])
</script>

<template>
  <section class="panel p-4">
    <header class="flex justify-between mb-2">
      <h2 class="font-display text-phosphor tracking-widest text-sm">报文 Hex</h2>
      <button v-if="store.err" type="button" class="text-danger text-xs" @click="store.err = ''">× 关闭错误</button>
    </header>
    <div v-if="store.selected" class="text-[10px] leading-5">
      <p class="text-ice mb-1">{{ store.selected.type }} · {{ store.selected.ts }}</p>
      <p class="break-all">
        <span v-for="(g, i) in groups" :key="i" class="hover:bg-phosphor/30 px-0.5" :title="`offset ${i}`">{{ g }} </span>
      </p>
    </div>
    <ul v-else class="max-h-28 overflow-auto space-y-1">
      <li v-for="p in store.packets.slice(-8).reverse()" :key="p.ts + p.type">
        <button type="button" class="text-left text-[10px] hover:text-phosphor" @click="store.selected = p">
          {{ p.type }} {{ (p.payload.flags as string[] | undefined)?.join('|') }} len={{ p.payload.len }}
        </button>
      </li>
      <li v-if="!store.packets.length" class="text-mist/60">空态：还没有精确包事件。</li>
    </ul>
  </section>
</template>
