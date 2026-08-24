<script setup lang="ts">
import { onMounted } from 'vue'
import { useTelemetry } from './stores/telemetry'
import StateWall from './components/StateWall.vue'
import WindowGrid from './components/WindowGrid.vue'
import TopologyFlow from './components/TopologyFlow.vue'
import ThroughputChart from './components/ThroughputChart.vue'
import ConnSidebar from './components/ConnSidebar.vue'
import ControlBar from './components/ControlBar.vue'
import HexInspector from './components/HexInspector.vue'

const store = useTelemetry()
onMounted(() => {
  store.connect()
  store.refresh()
  window.setInterval(() => store.refresh(), 2000)
})
</script>

<template>
  <div class="hexbg scanlines min-h-screen w-full">
    <header class="w-full border-b border-line px-5 py-3 flex flex-wrap items-center gap-4">
      <div>
        <p class="font-display text-xl text-phosphor tracking-[0.28em]">MINI GOSTACK</p>
        <p class="text-[11px] text-mist">用户态 TCP 雷达 · {{ store.device }} · {{ store.health }}</p>
      </div>
      <span
        class="ml-2 w-2.5 h-2.5 rounded-full"
        :class="store.wsState === 'on' ? 'bg-phosphor shadow-[0_0_8px_#39ff88]' : 'bg-amber animate-pulse'"
        :title="store.wsState"
      />
      <ControlBar class="ml-auto" />
    </header>

    <div
      v-if="store.toast"
      class="fixed top-4 right-4 z-[60] panel px-4 py-2 text-sm flex items-center gap-3"
      :class="store.err ? 'border-danger text-danger' : 'text-phosphor'"
    >
      <span>{{ store.toast }}</span>
      <button type="button" aria-label="关闭提示" @click="store.toast = ''; store.err = ''">×</button>
    </div>

    <main class="w-full p-4 grid gap-4 lg:grid-cols-[280px_1fr] xl:grid-cols-[280px_1fr_1fr]">
      <ConnSidebar class="min-h-[240px]" />
      <div class="space-y-4">
        <StateWall />
        <WindowGrid />
      </div>
      <div class="space-y-4 xl:block">
        <TopologyFlow />
        <ThroughputChart />
        <HexInspector />
      </div>
    </main>
  </div>
</template>
