#!/usr/bin/env sh
set -eu

if [ -z "${JWT_SECRET:-}" ] && [ -f .env ]; then
  JWT_SECRET="$(grep '^JWT_SECRET=' .env | cut -d= -f2-)"
  export JWT_SECRET
fi

API_URL="${API_URL:-http://localhost:5000}"
if [ -z "${JWT_SECRET:-}" ]; then
  echo "JWT_SECRET must be set in the environment or .env"
  exit 1
fi

STAMP="$(date +%s)"
USERNAME="tester_$STAMP"
OTHER_USERNAME="other_$STAMP"
PASSWORD="password123"

request() {
  method="$1"
  path="$2"
  body="${3:-}"
  token="${4:-}"

  if [ -n "$body" ] && [ -n "$token" ]; then
    curl -sS -X "$method" "$API_URL$path" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $token" \
      -d "$body" \
      -w '\n%{http_code}'
  elif [ -n "$body" ]; then
    curl -sS -X "$method" "$API_URL$path" \
      -H "Content-Type: application/json" \
      -d "$body" \
      -w '\n%{http_code}'
  elif [ -n "$token" ]; then
    curl -sS -X "$method" "$API_URL$path" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $token" \
      -w '\n%{http_code}'
  else
    curl -sS -X "$method" "$API_URL$path" \
      -H "Content-Type: application/json" \
      -w '\n%{http_code}'
  fi
}

assert_status() {
  response="$1"
  expected="$2"
  label="$3"

  status="$(printf '%s' "$response" | tail -n 1)"
  body="$(printf '%s' "$response" | sed '$d')"

  if [ "$status" != "$expected" ]; then
    echo "FAIL: $label"
    echo "Expected status: $expected"
    echo "Actual status:   $status"
    echo "Body:"
    echo "$body"
    exit 1
  fi

  echo "PASS: $label"
  printf '%s\n' "$body"
  echo
}

assert_code() {
  actual="$1"
  expected="$2"
  label="$3"

  if [ "$actual" != "$expected" ]; then
    echo "FAIL: $label"
    echo "Expected status: $expected"
    echo "Actual status:   $actual"
    exit 1
  fi

  echo "PASS: $label"
  echo
}

extract_token() {
  printf '%s' "$1" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p'
}

extract_todo_id() {
  printf '%s' "$1" | sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p'
}

expired_token() {
  python3 - "$JWT_SECRET" <<'PY'
import base64
import hashlib
import hmac
import json
import sys
import time

secret = sys.argv[1].encode()

def b64url(data):
    return base64.urlsafe_b64encode(data).rstrip(b"=").decode()

header = {"alg": "HS256", "typ": "JWT"}
payload = {
    "user_id": 999999,
    "username": "expired_user",
    "exp": int(time.time()) - 60,
    "iat": int(time.time()) - 120,
}

signing_input = ".".join([
    b64url(json.dumps(header, separators=(",", ":")).encode()),
    b64url(json.dumps(payload, separators=(",", ":")).encode()),
])
signature = hmac.new(secret, signing_input.encode(), hashlib.sha256).digest()
print(signing_input + "." + b64url(signature))
PY
}

echo "API_URL=$API_URL"
echo

HEALTH="$(request GET /health)"
assert_status "$HEALTH" 200 "health returns 200"

CORS_STATUS="$(curl -s -o /dev/null -w "%{http_code}" -X OPTIONS "$API_URL/todos" \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type, Authorization")"
assert_code "$CORS_STATUS" 204 "CORS preflight returns 204"

HEALTH_METHOD="$(request POST /health)"
assert_status "$HEALTH_METHOD" 405 "health rejects POST"

BAD_REGISTER_SHORT_USERNAME="$(request POST /auth/register '{"username":"ab","password":"password123"}')"
assert_status "$BAD_REGISTER_SHORT_USERNAME" 400 "register rejects short username"

BAD_REGISTER_SHORT_PASSWORD="$(request POST /auth/register '{"username":"validname","password":"short"}')"
assert_status "$BAD_REGISTER_SHORT_PASSWORD" 400 "register rejects short password"

BAD_REGISTER_JSON="$(request POST /auth/register '{"username":')"
assert_status "$BAD_REGISTER_JSON" 400 "register rejects malformed JSON"

REGISTER="$(request POST /auth/register "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")"
assert_status "$REGISTER" 201 "register creates user"
REGISTER_BODY="$(printf '%s' "$REGISTER" | sed '$d')"
TOKEN="$(extract_token "$REGISTER_BODY")"
if [ -z "$TOKEN" ]; then
  echo "FAIL: register response did not include token"
  exit 1
fi

DUPLICATE_REGISTER="$(request POST /auth/register "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")"
assert_status "$DUPLICATE_REGISTER" 409 "register rejects duplicate username"

BAD_LOGIN="$(request POST /auth/login "{\"username\":\"$USERNAME\",\"password\":\"wrongpass\"}")"
assert_status "$BAD_LOGIN" 401 "login rejects wrong password"

BAD_LOGIN_JSON="$(request POST /auth/login '{"username":')"
assert_status "$BAD_LOGIN_JSON" 400 "login rejects malformed JSON"

LOGIN="$(request POST /auth/login "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")"
assert_status "$LOGIN" 200 "login returns token"
LOGIN_BODY="$(printf '%s' "$LOGIN" | sed '$d')"
TOKEN="$(extract_token "$LOGIN_BODY")"
if [ -z "$TOKEN" ]; then
  echo "FAIL: login response did not include token"
  exit 1
fi

ME_NO_TOKEN="$(request GET /auth/me)"
assert_status "$ME_NO_TOKEN" 401 "me rejects missing token"

ME_BAD_TOKEN="$(request GET /auth/me "" "not-a-real-token")"
assert_status "$ME_BAD_TOKEN" 401 "me rejects invalid token"

EXPIRED_TOKEN="$(expired_token)"
ME_EXPIRED_TOKEN="$(request GET /auth/me "" "$EXPIRED_TOKEN")"
assert_status "$ME_EXPIRED_TOKEN" 401 "me rejects expired token"

ME="$(request GET /auth/me "" "$TOKEN")"
assert_status "$ME" 200 "me returns current user"

TODOS_NO_TOKEN="$(request GET /todos)"
assert_status "$TODOS_NO_TOKEN" 401 "todos rejects missing token"

TODOS_METHOD="$(request PATCH /todos "" "$TOKEN")"
assert_status "$TODOS_METHOD" 405 "todos rejects unsupported method"

CREATE_EMPTY="$(request POST /todos '{"title":""}' "$TOKEN")"
assert_status "$CREATE_EMPTY" 400 "create todo rejects empty title"

CREATE_BAD_JSON="$(request POST /todos '{"title":' "$TOKEN")"
assert_status "$CREATE_BAD_JSON" 400 "create todo rejects malformed JSON"

CREATE="$(request POST /todos '{"title":"Finish backend CRUD"}' "$TOKEN")"
assert_status "$CREATE" 201 "create todo succeeds"
CREATE_BODY="$(printf '%s' "$CREATE" | sed '$d')"
TODO_ID="$(extract_todo_id "$CREATE_BODY")"
if [ -z "$TODO_ID" ]; then
  echo "FAIL: create todo response did not include id"
  exit 1
fi
printf '%s' "$CREATE_BODY" | grep -q '"updated_at":' || {
  echo "FAIL: create todo response did not include updated_at"
  exit 1
}

LIST="$(request GET /todos "" "$TOKEN")"
assert_status "$LIST" 200 "list todos succeeds"

UPDATE_DONE="$(request PUT "/todos/$TODO_ID" '{"title":"Finish backend CRUD today","completed":true}' "$TOKEN")"
assert_status "$UPDATE_DONE" 200 "update todo can mark done and rename"
UPDATE_DONE_BODY="$(printf '%s' "$UPDATE_DONE" | sed '$d')"
printf '%s' "$UPDATE_DONE_BODY" | grep -q '"completed":true' || {
  echo "FAIL: update did not mark todo as completed"
  exit 1
}
printf '%s' "$UPDATE_DONE_BODY" | grep -q '"updated_at":' || {
  echo "FAIL: update response did not include updated_at"
  exit 1
}

UPDATE_UNDONE="$(request PUT "/todos/$TODO_ID" '{"title":"Finish backend CRUD today","completed":false}' "$TOKEN")"
assert_status "$UPDATE_UNDONE" 200 "update todo can unmark done"
UPDATE_UNDONE_BODY="$(printf '%s' "$UPDATE_UNDONE" | sed '$d')"
printf '%s' "$UPDATE_UNDONE_BODY" | grep -q '"completed":false' || {
  echo "FAIL: update did not unmark todo"
  exit 1
}

UPDATE_EMPTY_TITLE="$(request PUT "/todos/$TODO_ID" '{"title":"","completed":true}' "$TOKEN")"
assert_status "$UPDATE_EMPTY_TITLE" 400 "update todo rejects empty title"

UPDATE_BAD_JSON="$(request PUT "/todos/$TODO_ID" '{"title":' "$TOKEN")"
assert_status "$UPDATE_BAD_JSON" 400 "update todo rejects malformed JSON"

UPDATE_BAD_ID="$(request PUT /todos/not-a-number '{"title":"Bad","completed":true}' "$TOKEN")"
assert_status "$UPDATE_BAD_ID" 400 "update todo rejects invalid id"

TODO_BY_ID_METHOD="$(request GET "/todos/$TODO_ID" "" "$TOKEN")"
assert_status "$TODO_BY_ID_METHOD" 405 "todo by id rejects unsupported method"

UPDATE_MISSING_ID="$(request PUT /todos/999999 '{"title":"Missing","completed":true}' "$TOKEN")"
assert_status "$UPDATE_MISSING_ID" 404 "update todo returns 404 for missing todo"

OTHER_REGISTER="$(request POST /auth/register "{\"username\":\"$OTHER_USERNAME\",\"password\":\"$PASSWORD\"}")"
assert_status "$OTHER_REGISTER" 201 "register second user"
OTHER_BODY="$(printf '%s' "$OTHER_REGISTER" | sed '$d')"
OTHER_TOKEN="$(extract_token "$OTHER_BODY")"
if [ -z "$OTHER_TOKEN" ]; then
  echo "FAIL: second register response did not include token"
  exit 1
fi

OTHER_UPDATE="$(request PUT "/todos/$TODO_ID" '{"title":"Steal todo","completed":true}' "$OTHER_TOKEN")"
assert_status "$OTHER_UPDATE" 404 "second user cannot update first user's todo"

OTHER_DELETE="$(request DELETE "/todos/$TODO_ID" "" "$OTHER_TOKEN")"
assert_status "$OTHER_DELETE" 404 "second user cannot delete first user's todo"

DELETE_BAD_ID="$(request DELETE /todos/not-a-number "" "$TOKEN")"
assert_status "$DELETE_BAD_ID" 400 "delete todo rejects invalid id"

DELETE="$(request DELETE "/todos/$TODO_ID" "" "$TOKEN")"
assert_status "$DELETE" 204 "delete todo succeeds"

DELETE_AGAIN="$(request DELETE "/todos/$TODO_ID" "" "$TOKEN")"
assert_status "$DELETE_AGAIN" 404 "delete todo returns 404 after deletion"

echo "All backend tests passed."
