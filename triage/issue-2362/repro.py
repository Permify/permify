#!/usr/bin/env python3
import argparse
import json
import urllib.error
import urllib.request


SCHEMA = """entity user {}

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
"""


def main():
    parser = argparse.ArgumentParser(description="Live API reproduction for #2362")
    parser.add_argument("--port", type=int, choices=(3476, 3486), default=3476)
    args = parser.parse_args()
    base = f"http://localhost:{args.port}/v1/tenants/t1"

    def post(path, body):
        url = base + path
        request = urllib.request.Request(
            url, data=json.dumps(body).encode(),
            headers={"Content-Type": "application/json"}, method="POST",
        )
        try:
            with urllib.request.urlopen(request, timeout=30) as response:
                status, result = response.status, json.load(response)
        except urllib.error.HTTPError as error:
            status, result = error.code, json.loads(error.read())
        print(json.dumps({"method": "POST", "url": url, "body": body,
                          "status": status, "response": result}), flush=True)
        if status != 200:
            raise RuntimeError(f"{path}: HTTP {status}")
        return result

    version = post("/schemas/write", {"schema": SCHEMA})["schema_version"]
    tuples = []
    attributes = []
    for app, org in (("app1", "root"), ("app2", "child1")):
        tuples.append({
            "entity": {"type": "PublishedApplication", "id": app},
            "relation": "granted_top_to_down_org",
            "subject": {"type": "Organization", "id": org, "relation": ""},
        })
        attributes.append({
            "entity": {"type": "Organization", "id": org},
            "attribute": "org_id",
            "value": {"@type": "type.googleapis.com/base.v1.StringValue",
                      "data": org},
        })
    snap = post("/data/write", {
        "metadata": {"schema_version": version},
        "tuples": tuples, "attributes": attributes,
    })["snap_token"]
    metadata = {"schema_version": version, "snap_token": snap, "depth": 20}
    failures = []
    for parents, expected_ids in (
        (["root", "child1"], ["app1", "app2"]),
        (["child1"], ["app2"]),
        (["other_parent"], []),
        ([], []),
    ):
        common = {
            "metadata": metadata,
            "subject": {"type": "user", "id": "1", "relation": ""},
            "context": {"tuples": [], "attributes": [],
                        "data": {"parents": parents}},
        }
        for app in ("app1", "app2"):
            expected = ("CHECK_RESULT_ALLOWED" if app in expected_ids
                        else "CHECK_RESULT_DENIED")
            entity = {"type": "PublishedApplication", "id": app}
            check = post("/permissions/check", {
                **common, "entity": entity, "permission": "LIST_VIEW",
            })
            subject = post("/permissions/subject-permission", {
                **common, "entity": entity,
            })
            if check.get("can") != expected:
                failures.append(["check", parents, app, expected, check])
            if subject.get("results", {}).get("LIST_VIEW") != expected:
                failures.append(["subject-permission", parents, app,
                                 expected, subject])
        lookup = post("/permissions/lookup-entity", {
            **common, "entity_type": "PublishedApplication",
            "permission": "LIST_VIEW", "page_size": 100,
            "continuous_token": "",
        })
        if sorted(lookup.get("entity_ids", [])) != expected_ids:
            failures.append(["lookup-entity", parents, expected_ids, lookup])
        if lookup.get("continuous_token"):
            failures.append(["unexpected continuation", parents, lookup])
    print(json.dumps({"failures": failures}), flush=True)
    raise SystemExit(bool(failures))


if __name__ == "__main__":
    main()
