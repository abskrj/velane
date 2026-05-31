/**
 * k6 load test for the Velane invoke API.
 *
 * Run:
 *   k6 run scripts/load-test.js
 *
 * Override defaults via env vars:
 *   BASE_URL   — control-plane base URL    (default: http://localhost:8080)
 *   API_KEY    — vl_ API key               (required)
 *   TENANT     — tenant slug               (default: myorg)
 *   SNIPPET    — snippet slug to invoke    (default: text-stats)
 *   ENV        — deployment environment    (default: prod)
 *   VUS        — peak virtual users        (default: 20)
 *   DURATION   — steady-state duration     (default: 30s)
 */

import http from 'k6/http'
import { check, sleep } from 'k6'
import { Rate, Trend, Counter } from 'k6/metrics'

// ---------- config ----------
const BASE_URL = __ENV.BASE_URL  || 'http://localhost:8080'
const API_KEY  = __ENV.API_KEY   || 'vl_3607d09bcd625f5963f989d07c3a7bbb'
const TENANT   = __ENV.TENANT    || 'myorg'
const ENV      = __ENV.ENV       || 'prod'
const PEAK_VUS = parseInt(__ENV.VUS      || '20')
const DURATION = __ENV.DURATION  || '30s'

// Snippets to exercise in round-robin.
// Each entry: [slug, body]
const SCENARIOS = [
  [
    'text-stats',
    JSON.stringify({ text: 'The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs!' }),
  ],
  [
    'json-flatten',
    JSON.stringify({ data: { user: { name: 'Alice', address: { city: 'NYC', zip: '10001' } }, active: true } }),
  ],
  [
    'csv-parse',
    JSON.stringify({ csv: 'name,age,city\nAlice,30,NYC\nBob,25,LA\nCarol,35,Chicago' }),
  ],
]

// ---------- load profile ----------
// Ramp up → steady state → ramp down
export const options = {
  stages: [
    { duration: '10s', target: Math.ceil(PEAK_VUS * 0.25) }, // ramp up to 25 %
    { duration: '10s', target: Math.ceil(PEAK_VUS * 0.75) }, // ramp up to 75 %
    { duration: DURATION,                  target: PEAK_VUS }, // steady state
    { duration: '10s', target: 0 },                           // ramp down
  ],
  thresholds: {
    // 95th-percentile latency under 2 s
    'invoke_duration': ['p(95)<2000'],
    // Error rate under 1 %
    'invoke_errors': ['rate<0.01'],
  },
}

// ---------- custom metrics ----------
const invokeDuration = new Trend('invoke_duration', true)  // milliseconds
const invokeErrors   = new Rate('invoke_errors')
const invokeCount    = new Counter('invoke_count')

// ---------- headers (shared) ----------
const HEADERS = {
  'Authorization': `Bearer ${API_KEY}`,
  'Content-Type': 'application/json',
}

// ---------- VU entrypoint ----------
export default function () {
  // Round-robin across snippets so all three get exercised.
  const [slug, body] = SCENARIOS[__VU % SCENARIOS.length]
  const url = `${BASE_URL}/v1/invoke/${TENANT}/${slug}?env=${ENV}`

  const start = Date.now()
  const res = http.post(url, body, { headers: HEADERS })
  const elapsed = Date.now() - start

  invokeDuration.add(elapsed)
  invokeCount.add(1)

  const ok = check(res, {
    'status 200':       (r) => r.status === 200,
    'has output':       (r) => {
      try { return JSON.parse(r.body).output !== undefined } catch { return false }
    },
    'status completed': (r) => {
      try { return JSON.parse(r.body).status === 'completed' } catch { return false }
    },
  })

  invokeErrors.add(!ok)

  // Small think-time between requests per VU to avoid thundering-herd.
  sleep(0.1)
}

// ---------- summary hook ----------
export function handleSummary(data) {
  const d = data.metrics

  const p50  = d.invoke_duration?.values?.['p(50)']?.toFixed(1)  ?? 'n/a'
  const p95  = d.invoke_duration?.values?.['p(95)']?.toFixed(1)  ?? 'n/a'
  const p99  = d.invoke_duration?.values?.['p(99)']?.toFixed(1)  ?? 'n/a'
  const rps  = d.invoke_count?.values?.rate?.toFixed(1)           ?? 'n/a'
  const errs = ((d.invoke_errors?.values?.rate ?? 0) * 100).toFixed(2)
  const total = d.invoke_count?.values?.count ?? 0

  console.log(`
┌─────────────────────────────────────┐
│         Velane Invoke Load Test      │
├─────────────────────────────────────┤
│  Total invocations : ${String(total).padEnd(15)}│
│  Throughput        : ${String(rps + ' req/s').padEnd(15)}│
│  p50 latency       : ${String(p50 + ' ms').padEnd(15)}│
│  p95 latency       : ${String(p95 + ' ms').padEnd(15)}│
│  p99 latency       : ${String(p99 + ' ms').padEnd(15)}│
│  Error rate        : ${String(errs + ' %').padEnd(15)}│
└─────────────────────────────────────┘
`)
  return {}
}
