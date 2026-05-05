import http from 'k6/http';
import { check } from 'k6';
import { Trend, Counter } from 'k6/metrics';
import { SharedArray } from 'k6/data';
import papaparse from 'https://jslib.k6.io/papaparse/5.1.1/index.js';

const FIXTURE_PATH = __ENV.FIXTURES_FILE || '../seed/fixtures.sample.csv';
const fixtures = new SharedArray('fixtures', () => {
  const csv = open(FIXTURE_PATH);
  return papaparse.parse(csv, { header: true, skipEmptyLines: true }).data.map((r) => ({
    student_id: parseInt(r.student_id, 10),
    class_id: parseInt(r.class_id, 10),
    school_id: parseInt(r.school_id, 10)
  }));
});

const BASE = __ENV.BASE_URL || 'http://localhost:8080';
const TOKEN = __ENV.SESSION_TOKEN;
if (!TOKEN) throw new Error('SESSION_TOKEN env var required');
const HEADERS = { Authorization: `Bearer ${TOKEN}`, 'Content-Type': 'application/json' };

const tRoster = new Trend('roster_ms', true);
const errs = new Counter('app_errors');

const SCHOOLS = (() => {
  const seen = new Set();
  for (const f of fixtures) seen.add(f.school_id);
  return Array.from(seen);
})();

export const options = {
  scenarios: {
    roster_only: {
      executor: 'constant-vus',
      vus: parseInt(__ENV.VUS || '50', 10),
      duration: __ENV.DUR || '30s',
      exec: 'rosterReads'
    }
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    'roster_ms': ['p(95)<200']
  }
};

function pickRandom(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

export function rosterReads() {
  const school = pickRandom(SCHOOLS);
  const r = http.get(`${BASE}/v1/people/students?institutionId=${school}&limit=50&offset=0`, {
    headers: HEADERS,
    tags: { name: 'roster' }
  });
  tRoster.add(r.timings.duration);
  const body = r.json() || {};
  const ok = Array.isArray(body.students) && body.students.length > 0;
  if (!check(r, {
    'roster 200': (x) => x.status === 200,
    'roster has students': () => ok
  })) errs.add(1);
}
