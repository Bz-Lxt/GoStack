<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import * as d3 from 'd3'
import { useTelemetry } from '../stores/telemetry'

const store = useTelemetry()
const el = ref<HTMLDivElement | null>(null)

function render() {
  const root = el.value
  if (!root) return
  root.innerHTML = ''
  const data = store.series
  const w = 640, h = 180, m = { t: 10, r: 12, b: 22, l: 40 }
  const svg = d3.select(root).append('svg').attr('viewBox', `0 0 ${w} ${h}`).attr('class', 'w-full h-[180px]')
  if (!data.length) {
    svg.append('text').attr('x', w/2).attr('y', h/2).attr('text-anchor', 'middle').attr('fill', '#9bb0b8')
      .attr('font-size', 12).text('等待聚合快照…')
    return
  }
  const x = d3.scaleLinear().domain(d3.extent(data, (d) => d.t) as [number, number]).range([m.l, w - m.r])
  const y = d3.scaleLinear().domain([0, Math.max(1, ...data.map((d) => Math.max(d.thru, d.cwnd)))]).range([h - m.b, m.t])
  const line = (k: 'thru' | 'goodput' | 'cwnd') => d3.line<typeof data[0]>().x((d) => x(d.t)).y((d) => y(d[k]))
  svg.append('path').datum(data).attr('fill', 'none').attr('stroke', '#7ee8fa').attr('stroke-width', 1.4).attr('d', line('thru'))
  svg.append('path').datum(data).attr('fill', 'none').attr('stroke', '#39ff88').attr('stroke-width', 1.4).attr('d', line('goodput'))
  svg.append('path').datum(data).attr('fill', 'none').attr('stroke', '#f0a202').attr('stroke-width', 1.2).attr('d', line('cwnd'))
}

onMounted(render)
watch(() => store.series.length, render)
</script>

<template>
  <section class="panel p-4">
    <header class="flex justify-between mb-2">
      <h2 class="font-display text-phosphor tracking-widest text-sm">吞吐 / cwnd 时序</h2>
      <span class="text-[10px] text-mist">ice=wire · phosphor=goodput · amber=cwnd</span>
    </header>
    <div ref="el" />
  </section>
</template>
