<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import * as d3 from 'd3'
import { useTelemetry } from '../stores/telemetry'

const store = useTelemetry()
const el = ref<HTMLDivElement | null>(null)

const nodes = [
  { id: 'CLOSED', x: 80, y: 160 },
  { id: 'LISTEN', x: 80, y: 50 },
  { id: 'SYN_SENT', x: 240, y: 50 },
  { id: 'SYN_RCVD', x: 240, y: 160 },
  { id: 'ESTABLISHED', x: 430, y: 105 },
  { id: 'FIN_WAIT_1', x: 620, y: 40 },
  { id: 'FIN_WAIT_2', x: 800, y: 40 },
  { id: 'CLOSING', x: 620, y: 120 },
  { id: 'TIME_WAIT', x: 800, y: 120 },
  { id: 'CLOSE_WAIT', x: 620, y: 200 },
  { id: 'LAST_ACK', x: 800, y: 200 },
]

const edges = [
  ['CLOSED', 'LISTEN'], ['CLOSED', 'SYN_SENT'], ['LISTEN', 'SYN_RCVD'],
  ['SYN_SENT', 'ESTABLISHED'], ['SYN_RCVD', 'ESTABLISHED'],
  ['ESTABLISHED', 'FIN_WAIT_1'], ['ESTABLISHED', 'CLOSE_WAIT'],
  ['FIN_WAIT_1', 'FIN_WAIT_2'], ['FIN_WAIT_1', 'CLOSING'], ['FIN_WAIT_1', 'TIME_WAIT'],
  ['FIN_WAIT_2', 'TIME_WAIT'], ['CLOSING', 'TIME_WAIT'],
  ['CLOSE_WAIT', 'LAST_ACK'], ['LAST_ACK', 'CLOSED'], ['TIME_WAIT', 'CLOSED'],
]

function draw() {
  const root = el.value
  if (!root) return
  root.innerHTML = ''
  const svg = d3.select(root).append('svg').attr('viewBox', '0 0 900 250').attr('class', 'w-full h-[250px]')
  svg.append('defs').append('linearGradient').attr('id', 'flow').attr('x1', '0').attr('x2', '1')
    .selectAll('stop').data(['#39ff88', '#7ee8fa']).enter().append('stop')
    .attr('offset', (_d: string, i: number) => i).attr('stop-color', (d: string) => d)
  const pos = Object.fromEntries(nodes.map((n) => [n.id, n]))
  edges.forEach(([a, b]) => {
    const A = pos[a], B = pos[b]
    svg.append('line').attr('x1', A.x).attr('y1', A.y).attr('x2', B.x).attr('y2', B.y)
      .attr('stroke', '#1c2a33').attr('stroke-width', 1.4).attr('data-edge', `${a}->${b}`)
  })
  const g = svg.selectAll('g.node').data(nodes).enter().append('g')
    .attr('transform', (d) => `translate(${d.x},${d.y})`)
  g.append('rect').attr('x', -62).attr('y', -16).attr('width', 124).attr('height', 32).attr('rx', 4)
    .attr('fill', '#0c1218').attr('stroke', '#1c2a33').attr('class', 'tile')
  g.append('text').attr('text-anchor', 'middle').attr('dy', 5)
    .attr('fill', '#9bb0b8').attr('font-size', 11).attr('font-family', 'Chakra Petch')
    .text((d) => d.id)
  paint()
}

function paint() {
  const root = el.value
  if (!root) return
  const svg = d3.select(root).select('svg')
  const cur = store.lastState
  svg.selectAll('g').each(function (d: any) {
    const on = d.id === cur
    d3.select(this).select('.tile')
      .attr('stroke', on ? '#39ff88' : '#1c2a33')
      .attr('fill', on ? 'rgba(57,255,136,0.12)' : '#0c1218')
    d3.select(this).select('text').attr('fill', on ? '#39ff88' : '#9bb0b8')
  })
  const last = store.transitions.at(-1)
  if (last) {
    svg.selectAll('line').attr('stroke', '#1c2a33')
    svg.select(`line[data-edge="${last.from}->${last.to}"]`).attr('stroke', '#39ff88').attr('stroke-width', 2)
  }
}

onMounted(draw)
watch(() => store.lastState, paint)
watch(() => store.transitions.length, paint)
</script>

<template>
  <section class="panel p-4">
    <header class="flex items-center justify-between mb-2">
      <h2 class="font-display text-phosphor tracking-widest text-sm">TCP 状态机翻转墙</h2>
      <span class="text-[11px] text-ice">{{ store.lastTrigger || '等待握手' }}</span>
    </header>
    <div ref="el" class="min-h-[250px]" />
    <p class="text-[10px] text-mist/70 mt-1">11 态 RFC 793 · 当前 {{ store.lastState }} · 已记录 {{ store.transitions.length }} 次跃迁</p>
  </section>
</template>
