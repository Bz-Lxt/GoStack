export type Envelope<T> = { ok: boolean; data: T; error: { code: string; message: string } | null }

export type ConnSnap = {
  id: string
  local: string
  remote: string
  state: string
  alive_ms: number
  rx_bytes: number
  tx_bytes: number
  cwnd: number
  rwnd: number
  rto_ms: number
  srtt_ms?: number
  una?: number
  nxt?: number
  phase?: string
}

export type Event = {
  v: number
  ts: string
  type: string
  conn_id: string
  precise: boolean
  payload: Record<string, unknown>
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const r = await fetch(path, {
    headers: { 'Content-Type': 'application/json', ...(init?.headers || {}) },
    ...init,
  })
  const j = (await r.json()) as Envelope<T>
  if (!j.ok) throw new Error(j.error?.message || 'request failed')
  return j.data
}

export const api = {
  health: () => req<{ status: string; time: string; device: string; tz: string; mode: string }>('/api/v1/health'),
  connections: () => req<ConnSnap[]>('/api/v1/connections'),
  stats: () => req<Record<string, number>>('/api/v1/stack/stats'),
  step: (enabled: boolean, pps: number) =>
    req('/api/v1/control/step-mode', { method: 'POST', body: JSON.stringify({ enabled, pps }) }),
  impair: (loss: number, delay_ms: number, jitter_ms: number, reorder: number) =>
    req('/api/v1/control/impair', { method: 'POST', body: JSON.stringify({ loss, delay_ms, jitter_ms, reorder }) }),
  traffic: (bytes = 65536) =>
    req('/api/v1/traffic/start', { method: 'POST', body: JSON.stringify({ bytes, scenario: 'sliding-window' }) }),
}
