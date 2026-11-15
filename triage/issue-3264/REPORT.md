# Triage #3264

Relates to #3264.

**Verdict:** confirmed  
**Confidence:** high

## Environment

The latest Compose instance was already running when checked; the reported
Compose instance was started for this investigation. Both used Postgres 16
via the supplied Compose configuration. Two additional instances were started
with the memory engine because the issue specifies memory. All four were
healthy; no authentication was configured and the existing `t1` tenant was
used. The original issue text did not include its purported Node.js script, so
the reproduction below implements the two reported cases and an ordering
control independently.

Image identities were taken from each running container's image ID, image
inspection and `permify version`, before the lookup reproduction:

| Instance | Tag | Digest and platform | `permify version` | Observation |
|---|---|---|---|---|
| Compose reported, Postgres, `localhost:3486` | `v1.7.5` | `ghcr.io/permify/permify@sha256:1ac0dea83d3942104d18683c47f39ff1acf41e8aa04c2dda963a7842889e4faa` (platform: `linux/amd64`) | `v1.7.5` | Both affected cases 100/221; control 221/221 |
| Memory reported, `localhost:3586` | `v1.7.5` | `ghcr.io/permify/permify@sha256:1ac0dea83d3942104d18683c47f39ff1acf41e8aa04c2dda963a7842889e4faa` (platform: `linux/amd64`) | `v1.7.5` | Both affected cases 100/221; control 221/221 |
| Compose latest, Postgres, `localhost:3476` | `latest` | `ghcr.io/permify/permify@sha256:dcee61e6a2819726369594c6fb23aabb3de484694f6f53435d26a2ab6fbe7974` (platform: `linux/amd64`) | `v1.7.6` | Both affected cases 100/221; control 221/221 |
| Memory latest, `localhost:3576` | `latest` | `ghcr.io/permify/permify@sha256:dcee61e6a2819726369594c6fb23aabb3de484694f6f53435d26a2ab6fbe7974` (platform: `linux/amd64`) | `v1.7.6` | Both affected cases 100/221; control 221/221 |

Pull an exact image and platform, for example:

```sh
docker pull --platform linux/amd64 ghcr.io/permify/permify@sha256:dcee61e6a2819726369594c6fb23aabb3de484694f6f53435d26a2ab6fbe7974
docker pull --platform linux/amd64 ghcr.io/permify/permify@sha256:1ac0dea83d3942104d18683c47f39ff1acf41e8aa04c2dda963a7842889e4faa
```

## Replication steps

Run on a disposable instance: writing the schema changes tenant `t1`. The
following specifies every request and tuple; `BASE` is the desired URL in the
table above. The `repro.py` script automates these steps using Python 3's
standard library. Run reported first, then latest, for example:

```sh
python3 triage/issue-3264/repro.py http://localhost:3586
python3 triage/issue-3264/repro.py http://localhost:3486
python3 triage/issue-3264/repro.py http://localhost:3576
python3 triage/issue-3264/repro.py http://localhost:3476
```

1. POST `BASE/v1/tenants/t1/schemas/write` with `Content-Type: application/json`
   and this JSON body (the schema is shown in full):

   ```json
   {"schema":"entity user {}\n\nentity organization {\n  relation member @user\n  relation parent @organization\n  permission view = member or parent.view\n}\n\nentity team {\n  relation member @user @team#member\n  permission view = member\n}\n"}
   ```

   Save the returned `schema_version` as `SCHEMA_VERSION`.

2. POST `BASE/v1/tenants/t1/data/write` with JSON body
   `{"metadata":{"schema_version":"SCHEMA_VERSION"},"tuples":[...]}`,
   replacing `SCHEMA_VERSION` with the returned value and `...` with exactly
   the following 663 tuples. Each row represents a tuple object in the form
   `{"entity":{"type":"ENTITY_TYPE","id":"ENTITY_ID"},"relation":"RELATION","subject":{"type":"SUBJECT_TYPE","id":"SUBJECT_ID","relation":"SUBJECT_RELATION"}}`;
   an empty subject relation is the empty string. The ranges denote one tuple
   for each zero-padded integer in the inclusive range.

   | Entity type | Entity ID | Relation | Subject type | Subject ID | Subject relation |
   |---|---|---|---|---|---|
   | organization | `a-root` | `member` | user | `first` | `""` |
   | organization | `z-root` | `member` | user | `last` | `""` |
   | team | `a-root` | `member` | user | `team-user` | `""` |
   | organization | `b-000` through `b-219` | `parent` | organization | `a-root` | `""` |
   | organization | `c-000` through `c-219` | `parent` | organization | `z-root` | `""` |
   | team | `b-000` through `b-219` | `member` | team | `a-root` | `member` |

   Save `snap_token` from the response, or use `""` if absent.

3. POST `BASE/v1/tenants/t1/permissions/lookup-entity` for each of the
   `(ENTITY_TYPE, USER)` pairs `("organization","first")`,
   `("organization","last")`, and `("team","team-user")`:

   ```json
   {"metadata":{"schema_version":"SCHEMA_VERSION","snap_token":"SNAP_TOKEN","depth":20},"entity_type":"ENTITY_TYPE","permission":"view","subject":{"type":"user","id":"USER"},"page_size":100,"continuous_token":""}
   ```

   Replace the placeholders with the values above. For each response, repeat
   this same POST with `continuous_token` set to the preceding response's
   token until it is empty; collect `entity_ids` from all pages. The token is
   opaque and should be copied from the response rather than reconstructed.

## Expected

Each lookup should return 221 distinct IDs exactly once. The first
organization and team lookups should include `a-root` and `b-000` through
`b-219`; the ordering control should include `z-root` and `c-000` through
`c-219`.

## Actual

On each instance, both the root-first organization and the team subject-set
lookup produced pages `[100, 0]`, totaling 100 unique IDs. The root-last
organization control produced `[100, 100, 21]`, totaling 221 unique IDs.
The first page of the affected lookups ran from `a-root` to `b-098`.

Representative response bodies observed on both memory instances for each
affected lookup, with the first page's ID list trimmed:

```json
{"entity_ids":["a-root","b-000","..."],"continuous_token":"Yi0wOTk="}
{"entity_ids":[],"continuous_token":""}
```

Thus the second page terminates normally despite 121 expected IDs not having
been returned. This verifies the symptom, not the proposed internal cause.

## Evidence

- `repro.py`: executable reproduction of the schema, all 663 tuples and all
  paginated requests described above.
- Compose logs for `permify-latest` included
  `time=2026-11-15T22:55:02.974Z level=INFO msg="finished call" protocol=grpc grpc.service=base.v1.Permission grpc.method=LookupEntity grpc.code=OK`.
- Compose logs for `permify-reported` included
  `time=2026-11-15T22:55:01.974Z level=INFO msg="finished call" protocol=grpc grpc.service=base.v1.Permission grpc.method=LookupEntity grpc.code=OK`.
- Memory logs likewise showed `grpc.method=LookupEntity grpc.code=OK` for
  both versions. These log excerpts are trimmed; the response bodies above,
  rather than the success status alone, establish the pagination behavior.
