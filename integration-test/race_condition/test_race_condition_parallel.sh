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
echo "=== AFTER DELETE (PARALLEL CHECKS) ==="
check_access 'read' &
CHECK_PID1=$!
check_access 'comment' &
CHECK_PID2=$!
check_access 'edit' &
CHECK_PID3=$!
wait $CHECK_PID1
wait $CHECK_PID2
wait $CHECK_PID3

echo ""
echo "=== AFTER 1s WAIT (PARALLEL CHECKS) ==="
sleep 1
check_access 'read' &
CHECK_PID4=$!
check_access 'comment' &
CHECK_PID5=$!
check_access 'edit' &
CHECK_PID6=$!
wait $CHECK_PID4
wait $CHECK_PID5
wait $CHECK_PID6

echo ""
echo "=== AFTER 5s WAIT (PARALLEL CHECKS) ==="
sleep 5
check_access 'read' &
CHECK_PID7=$!
check_access 'comment' &
CHECK_PID8=$!
check_access 'edit' &
CHECK_PID9=$!
wait $CHECK_PID7
wait $CHECK_PID8
wait $CHECK_PID9
