# Triage #2362: not-reproduced

Relates to #2362.

**Verdict:** not-reproduced

**Confidence:** high for the explicit reproduction below; the original schema is incomplete.

Lookup Entity agrees with Check and Subject Permission for positive, subset,
unrelated, and empty `context.data.parents` lists on both tested releases.
No production source was changed.

## Scope and assumptions

The original schema references undefined `granted_common_perm`. For this test,
`LIST_VIEW` references the defined `granted_top_to_down_perm` instead, preserving
the intermediate permission and the relation-to-rule traversal described in the
report. An explicit `user` entity supplies the checking subject.

The original report names no affected release. Live issue comments identify
`v1.4.6` as a release containing a fix, and the reporter says the preceding
nightly worked. We tested that released version as an additional comparison,
not as a claimed affected version. No older affected release was established,
so this is **not** a `fixed-in-latest` verdict.

The later comment about pagination with complex data is a separate symptom.
This two-entity reproduction does not establish whether that pagination problem
exists. Reproducing it would require the relevant schema, data, exact requests
and continuation-token responses, and affected release.

## Environment

Both instances use released images, REST, tenant `t1`, no authentication, and
Postgres via the existing Docker Compose configuration. No checkout source was
built or run. Latest was already running (`GET /healthz` returned
`{"status":"SERVING"}`); the comparison instance was started during this triage
and returned the same health response.

Identities were captured from the running containers **before** reproduction,
in this order for each instance: container `.Image`, inspection of that exact
local image ID for `.RepoDigests` and `.Os/.Architecture`, then execution of
`permify version` inside the container.

| Instance | Image tag | Running local image ID | Registry digest and platform | `permify version` |
| --- | --- | --- | --- | --- |
| `permify-latest` (already running) | `ghcr.io/permify/permify:latest` | `sha256:81e8f2fe7c796049bc0ab66768f03af555e5903020cf33fce5ef2f52178ac5af` | ghcr.io/permify/permify@sha256:dcee61e6a2819726369594c6fb23aabb3de484694f6f53435d26a2ab6fbe7974 (platform: linux/amd64) | `v1.7.6` |
| `permify-reported` (started here) | `ghcr.io/permify/permify:v1.4.6` | `sha256:42c447216c56f00be90e7de8233bcd47450e1ea06f0c3de311994fb382e1e495` | ghcr.io/permify/permify@sha256:0fce12974e2a736d4b05f712af5e8982af09e80f1c8ebb014357e3038e3837e4 (platform: linux/amd64) | `v1.4.6` |

Exact registry/platform output captured:

```text
[ghcr.io/permify/permify@sha256:dcee61e6a2819726369594c6fb23aabb3de484694f6f53435d26a2ab6fbe7974] linux/amd64
[ghcr.io/permify/permify@sha256:0fce12974e2a736d4b05f712af5e8982af09e80f1c8ebb014357e3038e3837e4] linux/amd64
```

For exact-image replay, pull the digest with its captured platform, not the
mutable tag:

```sh
docker pull --platform linux/amd64 ghcr.io/permify/permify@sha256:dcee61e6a2819726369594c6fb23aabb3de484694f6f53435d26a2ab6fbe7974
docker pull --platform linux/amd64 ghcr.io/permify/permify@sha256:0fce12974e2a736d4b05f712af5e8982af09e80f1c8ebb014357e3038e3837e4
```

## Replication steps

1. Use isolated triage databases: writing the schema updates tenant `t1`'s schema.
   The latest service was already available on port `3476`. Start the additional
   release and check readiness:

   ```sh
   REPORTED_VERSION=v1.4.6 docker compose -p triage -f /home/runner/work/permify/permify/.github/triage/docker-compose.yml --profile reported up -d db-reported permify-reported
   curl -sS http://localhost:3486/healthz
   ```

   Perform steps 2–5 against `BASE=http://localhost:3486/v1/tenants/t1` first,
   then repeat against `BASE=http://localhost:3476/v1/tenants/t1`. All requests
   below are JSON POSTs with `Content-Type: application/json`, without credentials.
   Substitute `BASE` in each URL. Each instance has its own schema/snapshot tokens.

2. `POST BASE/schemas/write` with `{"schema": "<full schema below>"}`, encoding
   the schema as a JSON string. Save the response's `schema_version` as `V`.

   ```perm
   entity user {}

   rule check_org_in_parent_tree(org_id string) {
     org_id in context.data.parents
   }

   entity Organization {
     relation parent @Organization
     attribute org_id string
     action TOP_TO_DOWN = check_org_in_parent_tree(org_id)
   }

   entity PublishedApplication {
     relation granted_top_to_down_org @Organization
     permission granted_top_to_down_perm = granted_top_to_down_org.TOP_TO_DOWN
     action LIST_VIEW = granted_top_to_down_perm
   }
   ```

3. `POST BASE/data/write` with the following complete data set. Replace `V`
   with the schema version from step 2. Save the response's `snap_token` as `S`.
   No other relationships or attributes are created by this reproduction.

   ```json
   {
     "metadata": {"schema_version": "V"},
     "tuples": [
       {
         "entity": {"type": "PublishedApplication", "id": "app1"},
         "relation": "granted_top_to_down_org",
         "subject": {"type": "Organization", "id": "root", "relation": ""}
       },
       {
         "entity": {"type": "PublishedApplication", "id": "app2"},
         "relation": "granted_top_to_down_org",
         "subject": {"type": "Organization", "id": "child1", "relation": ""}
       }
     ],
     "attributes": [
       {
         "entity": {"type": "Organization", "id": "root"},
         "attribute": "org_id",
         "value": {"@type": "type.googleapis.com/base.v1.StringValue", "data": "root"}
       },
       {
         "entity": {"type": "Organization", "id": "child1"},
         "attribute": "org_id",
         "value": {"@type": "type.googleapis.com/base.v1.StringValue", "data": "child1"}
       }
     ]
   }
   ```

4. For each `P` in `["root","child1"]`, `["child1"]`, `["other_parent"]`,
   and `[]`, execute these three request templates in order. Substitute `P`
   as a JSON array, `V` and `S` with the returned strings, and run the first
   two requests for both `APP=app1` and `APP=app2` before the lookup.
   This is exactly 20 permission requests per instance.

   **`POST BASE/permissions/check`**

   ```json
   {
     "metadata": {"schema_version": "V", "snap_token": "S", "depth": 20},
     "subject": {"type": "user", "id": "1", "relation": ""},
     "context": {"tuples": [], "attributes": [], "data": {"parents": "P"}},
     "entity": {"type": "PublishedApplication", "id": "APP"},
     "permission": "LIST_VIEW"
   }
   ```

   **`POST BASE/permissions/subject-permission`**

   ```json
   {
     "metadata": {"schema_version": "V", "snap_token": "S", "depth": 20},
     "subject": {"type": "user", "id": "1", "relation": ""},
     "context": {"tuples": [], "attributes": [], "data": {"parents": "P"}},
     "entity": {"type": "PublishedApplication", "id": "APP"}
   }
   ```

   **`POST BASE/permissions/lookup-entity`**

   ```json
   {
     "metadata": {"schema_version": "V", "snap_token": "S", "depth": 20},
     "subject": {"type": "user", "id": "1", "relation": ""},
     "context": {"tuples": [], "attributes": [], "data": {"parents": "P"}},
     "entity_type": "PublishedApplication",
     "permission": "LIST_VIEW",
     "page_size": 100,
     "continuous_token": ""
   }
   ```

5. Compare the check `can` and subject-permission `results.LIST_VIEW` fields
   against membership in lookup's `entity_ids`. The complete observed matrix,
   identical on both instances, is:

   | `P` | app1 (both check APIs) | app2 (both check APIs) | Lookup `entity_ids` | `continuous_token` |
   | --- | --- | --- | --- | --- |
   | `["root","child1"]` | `CHECK_RESULT_ALLOWED` | `CHECK_RESULT_ALLOWED` | `["app1","app2"]` | `""` |
   | `["child1"]` | `CHECK_RESULT_DENIED` | `CHECK_RESULT_ALLOWED` | `["app2"]` | `""` |
   | `["other_parent"]` | `CHECK_RESULT_DENIED` | `CHECK_RESULT_DENIED` | `[]` | `""` |
   | `[]` | `CHECK_RESULT_DENIED` | `CHECK_RESULT_DENIED` | `[]` | `""` |

   Automated replay (Python 3 standard library only):

   ```sh
   python3 /home/runner/work/permify/permify/triage/issue-2362/repro.py --port 3486
   python3 /home/runner/work/permify/permify/triage/issue-2362/repro.py --port 3476
   ```

   The script prints every full request, HTTP status, and response. Both runs
   exited 0 with `{"failures": []}`. It asserts the expected decisions and empty
   continuation tokens; unexpected results cause a nonzero exit.

## Results by version

| Image tag | Digest and platform | `permify version` | Observed |
| --- | --- | --- | --- |
| `ghcr.io/permify/permify:v1.4.6` | ghcr.io/permify/permify@sha256:0fce12974e2a736d4b05f712af5e8982af09e80f1c8ebb014357e3038e3837e4 (platform: linux/amd64) | `v1.4.6` | 22/22 HTTP 200; all four contexts matched check and subject-permission; no missing entities. |
| `ghcr.io/permify/permify:latest` | ghcr.io/permify/permify@sha256:dcee61e6a2819726369594c6fb23aabb3de484694f6f53435d26a2ab6fbe7974 (platform: linux/amd64) | `v1.7.6` | 22/22 HTTP 200; all four contexts matched check and subject-permission; no missing entities. |

## Expected and actual

**Expected:** Lookup uses the supplied context to return exactly the applications
whose `LIST_VIEW` checks are allowed, including through the intermediate permission.

**Actual:** Matches expected on both releases. Example response bodies for
`parents=["root","child1"]`, with subject-permission trimmed to the relevant field:

```json
{"can":"CHECK_RESULT_ALLOWED","metadata":{"check_count":4}}
{"results":{"LIST_VIEW":"CHECK_RESULT_ALLOWED"}}
{"entity_ids":["app1","app2"],"continuous_token":""}
```

For `parents=["child1"]`, app1 was denied, app2 allowed, and lookup returned:

```json
{"entity_ids":["app2"],"continuous_token":""}
```

For both negative contexts, both applications were denied and lookup returned:

```json
{"entity_ids":[],"continuous_token":""}
```

## Evidence

Reproduction script: `/home/runner/work/permify/permify/triage/issue-2362/repro.py`.
It creates the full schema/data above and checks all three APIs, without running
or importing repository source.

Logs were collected with:

```sh
REPORTED_VERSION=v1.4.6 docker compose -p triage -f /home/runner/work/permify/permify/.github/triage/docker-compose.yml --profile reported logs --no-color --tail 20 permify-reported permify-latest
```

Key captured lines (transport fields omitted):

```text
permify-reported-1 | time=2026-11-17T15:07:06.754Z level=INFO msg="finished call" grpc.service=base.v1.Permission grpc.method=Check grpc.code=OK
permify-reported-1 | time=2026-11-17T15:07:06.756Z level=INFO msg="finished call" grpc.service=base.v1.Permission grpc.method=SubjectPermission grpc.code=OK
permify-reported-1 | time=2026-11-17T15:07:06.778Z level=INFO msg="finished call" grpc.service=base.v1.Permission grpc.method=LookupEntity grpc.code=OK
permify-latest-1 | time=2026-11-17T15:07:12.348Z level=INFO msg="finished call" grpc.service=base.v1.Permission grpc.method=Check grpc.code=OK
permify-latest-1 | time=2026-11-17T15:07:12.349Z level=INFO msg="finished call" grpc.service=base.v1.Permission grpc.method=SubjectPermission grpc.code=OK
permify-latest-1 | time=2026-11-17T15:07:12.366Z level=INFO msg="finished call" grpc.service=base.v1.Permission grpc.method=LookupEntity grpc.code=OK
```

These lines establish successful calls, not authorization outcomes; the response
bodies and script assertions establish those.
