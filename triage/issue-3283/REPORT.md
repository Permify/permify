# Triage #3283

Relates to #3283

**Verdict:** inconclusive  
**Confidence:** low

## Environment

- `ghcr.io/permify/permify:latest` was already running and healthy on first
  inspection. It was recreated from the same local image to enable garbage
  collection. Immediately after recreation the running container reported
  image ID `sha256:81e8f2fe7c796049bc0ab66768f03af555e5903020cf33fce5ef2f52178ac5af`,
  image `ghcr.io/permify/permify@sha256:dcee61e6a2819726369594c6fb23aabb3de484694f6f53435d26a2ab6fbe7974`
  (platform: `linux/amd64`), and `permify version` output `v1.7.6`.
- PostgreSQL `16.16`, via the triage Docker Compose database. No separate
  reported-version instance was started: the issue specifies no version.
- The test used 100,000 organization **entities in one tenant**, not 100,000
  tenants. The issue does not specify whether its organizations are tenants.
  The collector interval was shortened from `1200s` to `120s`, and its window
  from `1200s` to `1s`, to observe fresh data; timeout was `5m`.
- No pool-acquisition-rate or error-rate metric was captured. The same
  PostgreSQL backend PID for these statements establishes reuse for the
  observed statements only; it does not rule out failures at other scales or
  under the reporter's deployment conditions.

## Replication steps

Use a **disposable** database: the schema write replaces tenant `t1`'s schema.
Run from the repository root. The following is the workload exercised; the
included `repro.sh` automates it and waits for two collections in one interval.

1. Start the database and recreate the released service with the following
   temporary Compose override:

   ```sh
   cat > /tmp/permify-3283-override.yml <<'EOF'
   services:
     permify-latest:
       environment:
         PERMIFY_DATABASE_GARBAGE_COLLECTION_ENABLED: "true"
         PERMIFY_DATABASE_GARBAGE_COLLECTION_INTERVAL: "120s"
         PERMIFY_DATABASE_GARBAGE_COLLECTION_WINDOW: "1s"
         PERMIFY_DATABASE_GARBAGE_COLLECTION_TIMEOUT: "5m"
   EOF
   docker compose -p triage -f .github/triage/docker-compose.yml up -d db-latest
   docker compose -p triage -f .github/triage/docker-compose.yml -f /tmp/permify-3283-override.yml up -d --no-deps --force-recreate permify-latest
   id=$(docker inspect --format '{{.Image}}' "$(docker compose -p triage -f .github/triage/docker-compose.yml ps -q permify-latest)")
   docker image inspect "$id" --format '{{.RepoDigests}} {{.Os}}/{{.Architecture}}'
   docker compose -p triage -f .github/triage/docker-compose.yml exec -T permify-latest permify version
   curl -sS http://localhost:3476/healthz
   ```

   Wait until the health response is `{"status":"SERVING"}`.

2. POST `http://localhost:3476/v1/tenants/t1/schemas/write` with JSON:

   ```json
   {"schema":"entity user {}\nentity organization {\n  relation member @user\n  permission view = member\n}\n"}
   ```

   The full schema is:

   ```perm
   entity user {}
   entity organization {
     relation member @user
     permission view = member
   }
   ```

3. POST `http://localhost:3476/v1/tenants/t1/data/write` 100 times. Each
   request contains 1,000 tuples of the following form, with `N` taking each
   integer from `0` through `99999` exactly once; there are no attributes:

   ```json
   {"metadata":{},"tuples":[{"entity":{"type":"organization","id":"orgN"},"relation":"member","subject":{"type":"user","id":"u1"}}]}
   ```

   For example, the first tuple uses `"id":"org0"` and the final one uses
   `"id":"org99999"`. To generate and send the exact 100 request bodies:

   ```sh
   python3 - <<'PY'
   import json
   import urllib.request
   url = "http://localhost:3476/v1/tenants/t1/data/write"
   for start in range(0, 100000, 1000):
       tuples = [{"entity": {"type": "organization", "id": f"org{i}"},
                  "relation": "member",
                  "subject": {"type": "user", "id": "u1"}}
                 for i in range(start, start + 1000)]
       request = urllib.request.Request(url, data=json.dumps({"metadata": {}, "tuples": tuples}).encode(),
                                        headers={"Content-Type": "application/json"}, method="POST")
       with urllib.request.urlopen(request, timeout=120) as response:
           response.read()
   PY
   ```

4. POST `http://localhost:3476/v1/tenants/t1/data/delete` with JSON:

   ```json
   {"tuple_filter":{"entity":{"type":"organization","id":"org0"},"relation":"member","subject":{"type":"user","ids":["u1"]}},"attribute_filter":{}}
   ```

   For the two smaller requests above, for example use `curl -H
   'Content-Type: application/json' -d '<JSON body>' '<URL>'`.

5. After writes complete, enable PostgreSQL statement logging, wait 250 seconds
   for two collections, and reset logging:

   ```sh
   docker compose -p triage -f .github/triage/docker-compose.yml exec -T db-latest psql -U postgres -d permify -v ON_ERROR_STOP=1 -c "ALTER SYSTEM SET log_statement = 'all';" -c "SELECT pg_reload_conf();"
   sleep 250
   docker compose -p triage -f .github/triage/docker-compose.yml exec -T db-latest psql -U postgres -d permify -v ON_ERROR_STOP=1 -c "ALTER SYSTEM RESET log_statement;" -c "SELECT pg_reload_conf();"
   docker compose -p triage -f .github/triage/docker-compose.yml logs --no-color permify-latest | grep 'Garbage collection'
   docker compose -p triage -f .github/triage/docker-compose.yml logs --no-color db-latest | grep -E 'execute .*DELETE FROM (relation_tuples|attributes|transactions)'
   ```

   In the observed run, logging was enabled for 130 seconds, reset, then
   enabled for a further 130 seconds and reset. The script's continuous
   250-second observation covers at least two 120-second ticks after the writes.

## Results by version

| Image tag | Running image digest and platform | `permify version` | Observed |
|---|---|---|---|
| `ghcr.io/permify/permify:latest` | `ghcr.io/permify/permify@sha256:dcee61e6a2819726369594c6fb23aabb3de484694f6f53435d26a2ab6fbe7974` (platform: `linux/amd64`) | `v1.7.6` | Two successful collection logs; all six observed DELETEs used PostgreSQL PID 275. No acquisition/error-rate metric available. |

Pull the tested platform with:

```sh
docker pull --platform linux/amd64 ghcr.io/permify/permify@sha256:dcee61e6a2819726369594c6fb23aabb3de484694f6f53435d26a2ab6fbe7974
```

## Expected and actual

**Expected:** The issue expects connection reuse for collector DELETEs, without
the reported 25,000 pool acquisitions per second and 92% errors. It also shows
the transaction cutoff as a SQL bind parameter.

**Actual:** The schema write returned
`{"schema_version":"dbt36ekimoes7394hh00"}`; all 100 data-write requests
completed, and the delete returned `{"snap_token":"ODg0Ojg4NDo4ODU6ODg0"}`.
The collector logged success twice. PostgreSQL logged all three DELETEs on
PID 275 in both observed cycles. The transaction cutoff appeared as the SQL
literal `'884'::xid8`, rather than the parameter shown in the issue. This
observation alone does not establish the reported connection or error rates.
After the second cycle, the database had one tenant, 100,000 tuple rows and
one transaction row; this test does not establish that the expired tuple was
physically removed.

## Evidence

- `repro.sh`: schema, 100,000-entity write workload, delete, logging, and
  collector observation.
- Permify logs:

  ```text
  time=2026-11-15T22:35:30.527Z level=INFO msg="Garbage collection completed successfully"
  time=2026-11-15T22:37:30.525Z level=INFO msg="Garbage collection completed successfully"
  ```

- PostgreSQL statement logs (statement cache identifiers omitted):

  ```text
  2026-11-15 22:35:30.525 UTC [275] LOG: execute ... DELETE FROM relation_tuples WHERE tenant_id = $1 AND expired_tx_id <> $2::xid8 AND expired_tx_id < $3::xid8
  2026-11-15 22:35:30.526 UTC [275] LOG: execute ... DELETE FROM attributes WHERE tenant_id = $1 AND expired_tx_id <> $2::xid8 AND expired_tx_id < $3::xid8
  2026-11-15 22:35:30.526 UTC [275] LOG: execute ... DELETE FROM transactions WHERE tenant_id = $1 AND id < '884'::xid8
  2026-11-15 22:37:30.524 UTC [275] LOG: execute ... DELETE FROM relation_tuples WHERE tenant_id = $1 AND expired_tx_id <> $2::xid8 AND expired_tx_id < $3::xid8
  2026-11-15 22:37:30.524 UTC [275] LOG: execute ... DELETE FROM attributes WHERE tenant_id = $1 AND expired_tx_id <> $2::xid8 AND expired_tx_id < $3::xid8
  2026-11-15 22:37:30.525 UTC [275] LOG: execute ... DELETE FROM transactions WHERE tenant_id = $1 AND id < '884'::xid8
  ```

PostgreSQL's only `ERROR`/`FATAL` line in the captured container log was
`relation "migrations" does not exist` during initial startup, before the
successful migration and before this test; it is not evidence of a collector
error. To make a closer comparison, the reporter would need to provide the
Permify version, whether organizations are separate tenants, pool and proxy
configuration, workload/transaction distribution, and the names and labels of
the metrics behind the reported acquisition and error rates.
