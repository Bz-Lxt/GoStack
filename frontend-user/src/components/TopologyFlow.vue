<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import * as d3 from 'd3'
import { useTelemetry } from '../stores/telemetry'

const store = useTelemetry()
const el = ref<HTMLDivElement | null>(null)
const layers = ['TUN', 'IP', 'CKSUM', 'TCP', 'STATE', 'BUF', 'APP']

function spawn(layer: string, drop = false) {
  const root = el.value
  if (!root) return
  const svg = d3.select(root).select('svg')
  const i = Math.max(0, layers.indexOf(layer.toUpperCase().slice(0, 5)))
  const idx = layers.findIndex((l) => layer.toUpperCase().includes(l) || l.includes(layer.toUpperCase()))
  const x0 = 40
  const x1 = 40 + (idx >= 0 ? idx : i) * 86
  const y = 36
  const g = svg.append('circle').attr('cx', x0).attr('cy', y).attr('r', 4)
    .attr('fill', drop ? '#ff4d6a' : '#7ee8fa')
  g.transition().duration(800).attr('cx', drop ? x1 : 40 + 6 * 86)
    .on('end', function () { d3.select(this).remove() })
}

onMounted(() => {
  const root = el.value
  if (!root) return
  const svg = d3.select(root).append('svg').attr('viewBox', '0 0 640 72').attr('class', 'w-full h-[72px]')
  layers.forEach((l, i) => {
    const x = 40 + i * 86
    svg.append('rect').attr('x', x - 28).attr('y', 18).attr('width', 56).attr('height', 36).attr('rx', 4)
      .attr('fill', '#0c1218').attr('stroke', '#1c2a33')
    svg.append('text').attr('x', x).attr('y', 40).attr('text-anchor', 'middle')
      .attr('fill', '#7ee8fa').attr('font-size', 10).attr('font-family', 'Chakra Petch').text(l)
  })
})

watch(() => store.packets.length, () => {
  const ev = store.packets.at(-1)
  if (!ev) return
  const layer = String(ev.payload.layer || 'tcp')
  spawn(layer, ev.type === 'checksum.error')
})
</script>

<template>
  <section class="panel p-4">
    <h2 class="font-display text-phosphor tracking-widest text-sm mb-2">协议栈流转拓扑</h2>
    <div ref="el" />
    <p class="text-[10px] text-mist/70">粒子 = 报文。校验失败在对应层炸开为红点。</p>
  </section>
</template>
