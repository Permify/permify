#!/bin/bash
set -e

# Write schema
curl -sS -w '\n' 'http://localhost:3477/v1/tenants/t1/schemas/write' \
     -H "Content-Type: application/json" \
     -d "$(cat "schema.perm" | jq -Rs '{schema: . }')"

# Grant permissions
curl -sS -w '\n' 'http://localhost:3477/v1/tenants/t1/data/write' \
     -H "Content-Type: application/json" \
     -d '{
         "metadata": {},
         "tuples": [
             {"entity": {"type": "doc", "id": "d1"}, "relation": "can_read", "subject": {"type": "user", "id": "u1"}},
             {"entity": {"type": "doc", "id": "d1"}, "relation": "can_comment", "subject": {"type": "user", "id": "u1"}},
             {"entity": {"type": "doc", "id": "d1"}, "relation": "can_edit", "subject": {"type": "user", "id": "u1"}}
         ]
     }'

# Define functions
function check_access() {
    ACTION=$1

    curl -sS -w '\n' 'http://localhost:3477/v1/tenants/t1/permissions/check' \
         -H "Content-Type: application/json" \
         -d "$(jq -Rn --arg action "$ACTION" '{
             metadata: {depth: 20},
             entity: {type: "doc", id: "d1"},
             permission: $action,
             subject: {type: "user", id: "u1"}
         }')"
}

function drop_access() {
    RELATION=$1

    curl -sS -w '\n' 'http://localhost:3477/v1/tenants/t1/data/delete' \
         -H "Content-Type: application/json" \
         -d "$(jq -Rn --arg relation "$RELATION" '{
             tuple_filter: {
                 entity: {type: "doc", id: "d1"},
                 relation: $relation,
                 subject: {type: "user", "ids": ["u1"]}
             },
             attribute_filter: {}
         }')"
}

echo "=== BEFORE DELETE ==="
echo "READ    = $(check_access 'read')"
echo "COMMENT = $(check_access 'comment')"
echo "EDIT    = $(check_access 'edit')"

echo ""
echo "=== CONCURRENT DELETE ==="
drop_access 'can_comment' &
PID1=$!
drop_access 'can_edit' &
PID2=$!
wait $PID1
wait $PID2

echo ""
echo "=== AFTER DELETE ==="
echo "READ    = $(check_access 'read')"
echo "COMMENT = $(check_access 'comment')"
echo "EDIT    = $(check_access 'edit')"

sleep 10

echo ""
echo "=== AFTER 10s WAIT ==="
echo "READ    = $(check_access 'read')"
echo "COMMENT = $(check_access 'comment')"
echo "EDIT    = $(check_access 'edit')"
