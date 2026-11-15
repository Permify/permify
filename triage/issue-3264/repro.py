#!/usr/bin/env python3
import json
import sys
import urllib.request

base = sys.argv[1].rstrip("/") if len(sys.argv) > 1 else "http://localhost:3586"
schema = """entity user {}

entity organization {
  relation member @user
  relation parent @organization
  permission view = member or parent.view
}

entity team {
  relation member @user @team#member
  permission view = member
}
"""


def post(path, body):
    request = urllib.request.Request(
        base + path,
        data=json.dumps(body).encode(),
        headers={"Content-Type": "application/json"},
    )
    with urllib.request.urlopen(request) as response:
        return json.load(response)


def tuple_(entity_type, entity_id, relation, subject_type, subject_id, subject_relation=""):
    return {
        "entity": {"type": entity_type, "id": entity_id},
        "relation": relation,
        "subject": {"type": subject_type, "id": subject_id, "relation": subject_relation},
    }


schema_version = post("/v1/tenants/t1/schemas/write", {"schema": schema})["schema_version"]
tuples = [
    tuple_("organization", "a-root", "member", "user", "first"),
    tuple_("organization", "z-root", "member", "user", "last"),
    tuple_("team", "a-root", "member", "user", "team-user"),
]
for i in range(220):
    tuples.extend([
        tuple_("organization", f"b-{i:03}", "parent", "organization", "a-root"),
        tuple_("organization", f"c-{i:03}", "parent", "organization", "z-root"),
        tuple_("team", f"b-{i:03}", "member", "team", "a-root", "member"),
    ])
write = post("/v1/tenants/t1/data/write", {
    "metadata": {"schema_version": schema_version},
    "tuples": tuples,
})

for entity_type, user, label in [
    ("organization", "first", "organization root sorts first"),
    ("organization", "last", "organization root sorts last"),
    ("team", "team-user", "team subject set"),
]:
    token = ""
    seen = []
    pages = []
    for _ in range(20):
        response = post("/v1/tenants/t1/permissions/lookup-entity", {
            "metadata": {
                "schema_version": schema_version,
                "snap_token": write.get("snap_token", ""),
                "depth": 20,
            },
            "entity_type": entity_type,
            "permission": "view",
            "subject": {"type": "user", "id": user},
            "page_size": 100,
            "continuous_token": token,
        })
        ids = response.get("entity_ids", [])
        token = response.get("continuous_token", "")
        seen.extend(ids)
        pages.append(len(ids))
        print(f"{label}: page={len(pages)} count={len(ids)} "
              f"first={ids[0] if ids else '-'} last={ids[-1] if ids else '-'} "
              f"token_present={bool(token)}")
        if not token:
            break
    else:
        raise RuntimeError(f"{label}: pagination did not end")
    print(f"{label}: returned={len(seen)} unique={len(set(seen))} expected=221 pages={pages}")
