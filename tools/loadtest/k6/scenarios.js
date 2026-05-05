import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';
import { SharedArray } from 'k6/data';
import papaparse from 'https://jslib.k6.io/papaparse/5.1.1/index.js';

const FIXTURE_PATH = __ENV.FIXTURES_FILE || '../seed/fixtures.csv';
const fixtures = new SharedArray('fixtures', () => {
  const csv = open(FIXTURE_PATH);
  return papaparse.parse(csv, { header: true, skipEmptyLines: true }).data.map((r) => ({
    student_id: parseInt(r.student_id, 10),
    class_id: parseInt(r.class_id, 10),
    school_id: parseInt(r.school_id, 10)
  }));
});

const meta = JSON.parse(open('../seed/fixtures.meta.json'));
const BASE = __ENV.BASE_URL || 'http://localhost:8080';
const TOKEN = __ENV.SESSION_TOKEN;
if (!TOKEN) throw new Error('SESSION_TOKEN env var required (login first, see README)');

const HEADERS = { Authorization: `Bearer ${TOKEN}`, 'Content-Type': 'application/json' };

const tRoster = new Trend('roster_ms', true);
const tIngest = new Trend('ingest_ms', true);
const tDash = new Trend('dashboard_ms', true);
const tDrill = new Trend('drilldown_ms', true);
const errs = new Counter('app_errors');

function uniqueSchools() {
  const seen = new Set();
  for (const f of fixtures) seen.add(f.school_id);
  return Array.from(seen);
}
const SCHOOLS = uniqueSchools();

function uniqueClasses() {
  const seen = new Set();
  for (const f of fixtures) seen.add(f.class_id);
  return Array.from(seen);
}
const CLASSES = uniqueClasses();

const studentsByClass = (() => {
  const m = new Map();
  for (const f of fixtures) {
    if (!m.has(f.class_id)) m.set(f.class_id, []);
    m.get(f.class_id).push(f.student_id);
  }
  return m;
})();

export const options = {
  scenarios: {
    A_roster_reads: {
      executor: 'constant-vus',
      vus: parseInt(__ENV.VUS_A || '50', 10),
      duration: __ENV.DUR_A || '60s',
      exec: 'rosterReads',
      tags: { scenario: 'A' },
      gracefulStop: '5s'
    },
    B_score_ingest: {
      executor: 'constant-vus',
      vus: parseInt(__ENV.VUS_B || '100', 10),
      duration: __ENV.DUR_B || '30s',
      exec: 'scoreIngest',
      startTime: '65s',
      tags: { scenario: 'B' },
      gracefulStop: '5s'
    },
    C_dashboard_reads: {
      executor: 'constant-vus',
      vus: parseInt(__ENV.VUS_C || '20', 10),
      duration: __ENV.DUR_C || '60s',
      exec: 'dashboardReads',
      startTime: '100s',
      tags: { scenario: 'C' },
      gracefulStop: '5s'
    },
    D_mixed: {
      executor: 'constant-vus',
      vus: parseInt(__ENV.VUS_D || '50', 10),
      duration: __ENV.DUR_D || '60s',
      exec: 'mixedTraffic',
      startTime: '165s',
      tags: { scenario: 'D' },
      gracefulStop: '5s'
    }
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    'roster_ms{scenario:A}': ['p(95)<2000'],
    'dashboard_ms{scenario:C}': ['p(95)<3000']
  }
};

function pickRandom(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

export function rosterReads() {
  const school = pickRandom(SCHOOLS);
  const url = `${BASE}/v1/people/students?institutionId=${school}&limit=50&offset=0`;
  const r = http.get(url, { headers: HEADERS, tags: { name: 'roster' } });
  tRoster.add(r.timings.duration);
  const body = r.json() || {};
  const ok = Array.isArray(body.students) && body.students.length > 0;
  if (!check(r, {
    'roster 200': (x) => x.status === 200,
    'roster has students': () => ok
  })) errs.add(1);
}

export function scoreIngest() {
  const classID = pickRandom(CLASSES);
  const studentList = studentsByClass.get(classID) || [];
  if (studentList.length === 0) return;
  const batch = studentList.slice(0, Math.min(30, studentList.length));
  const entries = batch.map((sid) => ({
    student_id: sid,
    raw_score: 20 + Math.floor(Math.random() * 80),
    mode: 'proctored_score'
  }));
  const url = `${BASE}/v1/teacher/classes/${classID}/campaigns/${meta.campaign_id}/scores`;
  const r = http.post(url, JSON.stringify({ entries }), { headers: HEADERS, tags: { name: 'ingest' } });
  tIngest.add(r.timings.duration);
  const body = r.json() || {};
  const wrote = (body.created || 0) + (body.updated || 0);
  if (!check(r, {
    'ingest 200': (x) => x.status === 200,
    'ingest wrote rows': () => wrote > 0
  })) errs.add(1);
}

export function dashboardReads() {
  const url = `${BASE}/v1/progression?scope_node_id=${meta.region_id}&period_id=${meta.period_id}&campaign_id=${meta.campaign_id}`;
  const r = http.get(url, { headers: HEADERS, tags: { name: 'dashboard' } });
  tDash.add(r.timings.duration);
  const body = r.json() || {};
  if (!check(r, {
    'dash 200': (x) => x.status === 200,
    'dash has bands': () => Array.isArray(body.bands) && body.bands.length > 0
  })) errs.add(1);
}

export function mixedTraffic() {
  const dice = Math.random();
  if (dice < 0.7) {
    scoreIngest();
  } else if (dice < 0.9) {
    rosterReads();
  } else {
    const url = `${BASE}/v1/progression/drilldown?scope_node_id=${meta.region_id}&period_id=${meta.period_id}&campaign_id=${meta.campaign_id}`;
    const r = http.get(url, { headers: HEADERS, tags: { name: 'drilldown' } });
    tDrill.add(r.timings.duration);
    if (!check(r, { 'drilldown 200': (x) => x.status === 200 })) errs.add(1);
  }
  sleep(0.1);
}
