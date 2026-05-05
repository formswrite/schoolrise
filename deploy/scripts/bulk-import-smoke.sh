#!/usr/bin/env bash
set -euo pipefail

GATEWAY="${GATEWAY:-http://localhost:8080}"
WEB="${WEB:-http://localhost:3001}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@local.test}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-ChangeMe123!}"
COUNT="${COUNT:-1500}"
TS=$(date +%s)
SUFFIX="bulk${TS}"

echo "==> Logging in as ${ADMIN_EMAIL}"
TOKEN=$(curl -fsS -X POST "${GATEWAY}/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASSWORD}\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['SessionToken'])")
[ -n "$TOKEN" ] || { echo "no token"; exit 1; }

echo "==> Creating hierarchy: region/prefecture/delegation/institution"
mk_node() {
  local label="$1" code="$2" level="$3" parent="${4:-null}"
  local payload
  if [ "$parent" = "null" ]; then
    payload=$(printf '{"code":"%s","label":"%s","level":"%s"}' "$code" "$label" "$level")
  else
    payload=$(printf '{"parentId":%s,"code":"%s","label":"%s","level":"%s"}' "$parent" "$code" "$label" "$level")
  fi
  curl -fsS -X POST "${GATEWAY}/v1/tenancy/nodes" \
    -H "Authorization: Bearer ${TOKEN}" -H "Content-Type: application/json" \
    -d "$payload" | python3 -c "import sys,json; print(json.load(sys.stdin)['node']['id'])"
}

REGION_ID=$(mk_node "Region ${SUFFIX}" "r-${SUFFIX}" "region")
INST_ID=$(mk_node   "School ${SUFFIX}" "i-${SUFFIX}" "institution" "$REGION_ID")
echo "    region=${REGION_ID} institution=${INST_ID}"

echo "==> Generating ${COUNT}-row CSV"
TMPCSV=$(mktemp)
echo "full_name,gender,date_of_birth,student_code,enrollment_date" > "$TMPCSV"
for i in $(seq 1 "$COUNT"); do
  G=$([ $((i % 2)) -eq 0 ] && echo "female" || echo "male")
  printf 'Student %05d %s,%s,2010-01-01,STU-%s-%05d,2025-09-01\n' "$i" "$SUFFIX" "$G" "$SUFFIX" "$i" >> "$TMPCSV"
done
echo "    csv: $(wc -l < "$TMPCSV") lines, $(stat -f%z "$TMPCSV" 2>/dev/null || stat -c%s "$TMPCSV") bytes"

echo "==> POST /v1/imports/students (dry_run=false)"
PAYLOAD=$(python3 -c "
import json, sys
with open('$TMPCSV') as f: csv = f.read()
print(json.dumps({'institution_id': $INST_ID, 'csv_data': csv, 'dry_run': False}))
")
START=$(date +%s)
RESP=$(curl -fsS -X POST "${GATEWAY}/v1/imports/students" \
  -H "Authorization: Bearer ${TOKEN}" -H "Content-Type: application/json" \
  -d "$PAYLOAD")
END=$(date +%s)
ELAPSED=$((END - START))
echo "    elapsed: ${ELAPSED}s"

python3 -c "
import json
r = json.loads('''$RESP''')
print(f'    job #{r[\"id\"]}: total={r[\"total_rows\"]} succ={r[\"succeeded\"]} fail={r[\"failed\"]} status={r[\"status\"]}')
assert r['succeeded'] == $COUNT, f'expected $COUNT successes, got {r[\"succeeded\"]}'
assert r['failed'] == 0, f'expected 0 failures, got {r[\"failed\"]}'
print('    ✓ all rows inserted')
"

echo "==> Coverage check"
COV=$(curl -fsS -H "Authorization: Bearer ${TOKEN}" \
  "${GATEWAY}/v1/enrollment/coverage?scope_node_id=${INST_ID}&period_id=1" || echo "{}")
echo "    coverage: $COV  (note: 0 enrolled — students are created but not enrolled by importer)"

rm -f "$TMPCSV"
echo "==> SMOKE OK"
