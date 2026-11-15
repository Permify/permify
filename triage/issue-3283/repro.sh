#!/usr/bin/env bash
set -euo pipefail

# Use only with a disposable triage database: this replaces tenant t1's schema.
cd "$(git rev-parse --show-toplevel)"
compose=(docker compose -p triage -f .github/triage/docker-compose.yml)
override=$(mktemp)
trap 'rm -f "$override"' EXIT
cat > "$override" <<'YAML'
services:
  permify-latest:
    environment:
      PERMIFY_DATABASE_GARBAGE_COLLECTION_ENABLED: "true"
      PERMIFY_DATABASE_GARBAGE_COLLECTION_INTERVAL: "120s"
      PERMIFY_DATABASE_GARBAGE_COLLECTION_WINDOW: "1s"
      PERMIFY_DATABASE_GARBAGE_COLLECTION_TIMEOUT: "5m"
YAML

"${compose[@]}" up -d db-latest
"${compose[@]}" -f "$override" up -d --no-deps --force-recreate permify-latest
image_id=$(docker inspect --format '{{.Image}}' "$("${compose[@]}" ps -q permify-latest)")
docker image inspect "$image_id" --format '{{.RepoDigests}} {{.Os}}/{{.Architecture}}'
"${compose[@]}" exec -T permify-latest permify version
for _ in $(seq 1 150); do
  if curl -fsS http://localhost:3476/healthz; then break; fi
  sleep 2
done
echo
curl -fsS http://localhost:3476/healthz
echo

python3 - <<'PY'
import json
import urllib.request

base = "http://localhost:3476/v1/tenants/t1/"

def post(path, body):
    request = urllib.request.Request(
        base + path, data=json.dumps(body).encode(),
        headers={"Content-Type": "application/json"}, method="POST",
    )
    with urllib.request.urlopen(request, timeout=120) as response:
        return response.read().decode()

print("schema:", post("schemas/write", {
    "schema": "entity user {}\nentity organization {\n  relation member @user\n  permission view = member\n}\n"
}), flush=True)
for start in range(0, 100000, 1000):
    tuples = [{
        "entity": {"type": "organization", "id": f"org{i}"},
        "relation": "member",
        "subject": {"type": "user", "id": "u1"},
    } for i in range(start, start + 1000)]
    post("data/write", {"metadata": {}, "tuples": tuples})
    if start % 10000 == 0:
        print(f"wrote {start + 1000} organizations", flush=True)
print("wrote 100000 organizations", flush=True)
print("delete:", post("data/delete", {
    "tuple_filter": {
        "entity": {"type": "organization", "id": "org0"},
        "relation": "member",
        "subject": {"type": "user", "ids": ["u1"]},
    },
    "attribute_filter": {},
}), flush=True)
PY

# Enable database statement logging only after the large writes.
"${compose[@]}" exec -T db-latest psql -U postgres -d permify -v ON_ERROR_STOP=1 \
  -c "ALTER SYSTEM SET log_statement = 'all';" -c "SELECT pg_reload_conf();"
echo "Waiting 250 seconds for two collections after writes and logging are enabled."
sleep 250
"${compose[@]}" exec -T db-latest psql -U postgres -d permify -v ON_ERROR_STOP=1 \
  -c "ALTER SYSTEM RESET log_statement;" -c "SELECT pg_reload_conf();"
echo "Collection logs:"
"${compose[@]}" logs --no-color permify-latest | grep 'Garbage collection' || true
echo "DELETE statements in PostgreSQL logs:"
"${compose[@]}" logs --no-color db-latest | grep -E 'execute .*DELETE FROM (relation_tuples|attributes|transactions)' || true
echo "PostgreSQL error log lines:"
"${compose[@]}" logs --no-color db-latest | grep -E ' ERROR:| FATAL:' || true
echo "Current database rows and connections:"
"${compose[@]}" exec -T db-latest psql -U postgres -d permify \
  -c "SELECT (SELECT count(*) FROM tenants) AS tenants, (SELECT count(*) FROM relation_tuples WHERE tenant_id='t1') AS tuples, (SELECT count(*) FROM transactions WHERE tenant_id='t1') AS transactions, (SELECT count(*) FROM pg_stat_activity WHERE datname='permify') AS connections;"
