#!/usr/bin/env bash
# Live proof of the two backend guarantees, against a running server.
#   Terminal 1:  cd backend && make run
#   Terminal 2:  ./demo.sh
set -euo pipefail

BASE="${BASE:-http://localhost:8080}"
BATCH="demo-$(date +%s)"
WORKERS='["w-001","w-002","w-003","w-004","w-005","w-006","w-007","w-008","w-009"]'

say() { printf '\n\033[1m%s\033[0m\n' "$1"; }

say "1 · Workers with pending disbursements"
curl -s "$BASE/workers" > /tmp/demo_workers.json
python3 - <<'PY'
import json
for w in json.load(open('/tmp/demo_workers.json')):
    print(f"  {w['id']}  {w['name']:<20} {w['amount']:>10} {w['currency']}")
PY

say "2 · Submit batch — returns 202 with everything pending"
curl -s -X POST "$BASE/disbursements" -H 'Content-Type: application/json' \
  -d "{\"batch_id\":\"$BATCH\",\"worker_ids\":$WORKERS}" > /tmp/demo_post.json
python3 - <<'PY'
import json
from collections import Counter
print("  ", dict(Counter(r["status"] for r in json.load(open('/tmp/demo_post.json'))["results"])))
PY

say "3 · Poll until settled (provider: 50-200ms latency, ~30% failures)"
for _ in $(seq 1 30); do
  curl -s "$BASE/disbursements/$BATCH" > /tmp/demo_first.json
  PENDING=$(python3 -c 'import json;print(sum(1 for r in json.load(open("/tmp/demo_first.json"))["results"] if r["status"]=="pending"))')
  [ "$PENDING" = "0" ] && break
done
python3 - <<'PY'
import json
from collections import Counter
print("  settled:", dict(Counter(r["status"] for r in json.load(open('/tmp/demo_first.json'))["results"])))
PY

say "4 · Idempotency — resubmit the SAME batch_id"
curl -s -X POST "$BASE/disbursements" -H 'Content-Type: application/json' \
  -d "{\"batch_id\":\"$BATCH\",\"worker_ids\":$WORKERS}" > /dev/null
curl -s "$BASE/disbursements/$BATCH" > /tmp/demo_second.json
python3 - <<'PY'
import json
a = json.load(open('/tmp/demo_first.json'))
b = json.load(open('/tmp/demo_second.json'))
txns = lambda x: [r.get("provider_txn_id") for r in x["results"]]
ok = a == b and txns(a) == txns(b)
print("  identical results, same txn ids, no double-pay:", "PASS ✅" if ok else "FAIL ❌")
PY
