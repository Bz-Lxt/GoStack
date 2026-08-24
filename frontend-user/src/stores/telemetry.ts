import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type ConnSnap, type Event } from '../lib/api'

export const STATES = [
  'CLOSED', 'LISTEN', 'SYN_SENT', 'SYN_RCVD', 'ESTABLISHED',
  'FIN_WAIT_1', 'FIN_WAIT_2', 'CLOSE_WAIT', 'CLOSING', 'LAST_ACK', 'TIME_WAIT',
] as const

export const useTelemetry = defineStore('telemetry', () => {
  const wsState = ref<'off' | 'on' | 'retry'>('off')
  const health = ref('')
  const device = ref('--')
  const conns = ref<ConnSnap[]>([])
  const activeId = ref('')
  const lastState = ref('CLOSED')
  const lastTrigger = ref('')
  const transitions = ref<{ from: string; to: string; at: string }[]>([])
  const cells = ref<{ seq: number; len: number; mark: string }[]>([])
  const packets = ref<Event[]>([])
  const selected = ref<Event | null>(null)
  const series = ref<{ t: number; goodput: number; thru: number; cwnd: number; rto: number; rwnd: number }[]>([])
  const toast = ref('')
  const err = ref('')
  let toastTimer: number | undefined

  function flash(msg: string, isErr = false) {
    if (isErr) err.value = msg
    toast.value = msg
    window.clearTimeout(toastTimer)
    toastTimer = window.setTimeout(() => {
      toast.value = ''
      err.value = ''
    }, 5000)
  }

  function ingest(ev: Event) {
    if (ev.conn_id && !activeId.value) activeId.value = ev.conn_id
    if (ev.type === 'state.transition') {
      const from = String(ev.payload.from || '')
      const to = String(ev.payload.to || '')
      lastState.value = to || lastState.value
      lastTrigger.value = String(ev.payload.trigger || '')
      transitions.value = [...transitions.value.slice(-40), { from, to, at: ev.ts }]
    }
    if (ev.type === 'window.update' && Array.isArray(ev.payload.cells)) {
      cells.value = ev.payload.cells as { seq: number; len: number; mark: string }[]
    }
    if (ev.type.startsWith('packet.') || ev.type === 'retransmit' || ev.type === 'checksum.error') {
      packets.value = [...packets.value.slice(-80), ev]
    }
    if (ev.type === 'aggregate.snapshot') {
      series.value = [
        ...series.value.slice(-120),
        {
          t: Date.now(),
          goodput: Number(ev.payload.goodput_bps || 0),
          thru: Number(ev.payload.throughput_bps || 0),
          cwnd: Number(ev.payload.cwnd || 0),
          rto: Number(ev.payload.rto_ms || 0),
          rwnd: Number(ev.payload.rwnd || 0),
        },
      ]
    }
  }

  function connect() {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const url = `${proto}://${location.host}/ws/telemetry`
    const ws = new WebSocket(url)
    wsState.value = 'retry'
    ws.onopen = () => { wsState.value = 'on' }
    ws.onmessage = (m) => {
      try { ingest(JSON.parse(m.data) as Event) } catch { /* ignore */ }
    }
    ws.onclose = () => {
      wsState.value = 'retry'
      window.setTimeout(connect, 1200)
    }
  }

  async function refresh() {
    try {
      const h = await api.health()
      health.value = h.time
      device.value = `${h.mode}/${h.device}`
      conns.value = await api.connections()
      if (!activeId.value && conns.value[0]) activeId.value = conns.value[0].id
    } catch (e) {
      flash((e as Error).message, true)
    }
  }

  return { wsState, health, device, conns, activeId, lastState, lastTrigger, transitions, cells, packets, selected, series, toast, err, flash, connect, refresh, ingest }
})
