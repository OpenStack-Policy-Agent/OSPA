#!/usr/bin/env bash
set -euo pipefail

# Creates 2 Nova servers:
# - ospa-e2e-compliant   (ACTIVE)
# - ospa-e2e-noncompliant (stopped to SHUTOFF)
#
# Requirements:
# - openstack CLI configured (clouds.yaml + OS_CLOUD)
# - quota to create 2 servers
#
# Usage:
#   export OS_CLOUD=mycloud
#   ./create_ospa_e2e_servers.sh
#
# Optional inputs:
#   export NAME_PREFIX=ospa-e2e
#   export IMAGE_NAME_OR_ID=...
#   export FLAVOR_NAME_OR_ID=...
#   export NETWORK_NAME_OR_ID=...
#   export KEY_NAME=...              # optional
#   export SECURITY_GROUP=...        # optional

: "${OS_CLOUD:?OS_CLOUD must be set (cloud name in clouds.yaml)}"

NAME_PREFIX="${NAME_PREFIX:-ospa-e2e}"
COMPLIANT_NAME="${NAME_PREFIX}-compliant"
NONCOMPLIANT_NAME="${NAME_PREFIX}-noncompliant"

pick_one() {
  # pick_one "<cmd that outputs ids>" "<human label>"
  local cmd="$1"
  local label="$2"
  local v
  v="$(bash -lc "$cmd" | head -n1 || true)"
  if [[ -z "${v}" ]]; then
    echo "ERROR: Could not auto-pick ${label}. Set the corresponding env var." >&2
    return 1
  fi
  echo "$v"
}

# Resolve image/flavor/network from env or auto-pick the first available.
IMAGE="${IMAGE_NAME_OR_ID:-}"
FLAVOR="${FLAVOR_NAME_OR_ID:-}"
NETWORK="${NETWORK_NAME_OR_ID:-}"

if [[ -z "$IMAGE" ]]; then
  IMAGE="$(pick_one "openstack image list -f value -c ID" "IMAGE_NAME_OR_ID")"
fi
if [[ -z "$FLAVOR" ]]; then
  FLAVOR="$(pick_one "openstack flavor list -f value -c ID" "FLAVOR_NAME_OR_ID")"
fi
if [[ -z "$NETWORK" ]]; then
  # Prefer a non-external network if available, else fall back to the first network.
  NETWORK="$(
    openstack network list -f value -c ID -c 'Router:External' 2>/dev/null \
      | awk '$2 != "True" {print $1; exit}' \
      || true
  )"
  if [[ -z "$NETWORK" ]]; then
    NETWORK="$(pick_one "openstack network list -f value -c ID" "NETWORK_NAME_OR_ID")"
  fi
fi

echo "Using:"
echo "  OS_CLOUD=$OS_CLOUD"
echo "  IMAGE=$IMAGE"
echo "  FLAVOR=$FLAVOR"
echo "  NETWORK=$NETWORK"
echo

create_server() {
  local name="$1"
  echo "Creating server: $name"
  local args=(openstack server create --image "$IMAGE" --flavor "$FLAVOR" --network "$NETWORK" "$name" -f value -c id)

  if [[ -n "${KEY_NAME:-}" ]]; then
    args+=(--key-name "$KEY_NAME")
  fi
  if [[ -n "${SECURITY_GROUP:-}" ]]; then
    args+=(--security-group "$SECURITY_GROUP")
  fi

  "${args[@]}"
}

wait_status() {
  local name="$1"
  local want="$2"
  echo "Waiting for $name to reach status=$want ..."
  local timeout_secs="${TIMEOUT_SECS:-600}" # 10 minutes default
  local start end got
  start="$(date +%s)"
  end="$((start + timeout_secs))"
  while true; do
    got="$(openstack server show "$name" -f value -c status 2>/dev/null || true)"
    if [[ "$got" == "$want" ]]; then
      return 0
    fi
    if [[ "$(date +%s)" -ge "$end" ]]; then
      echo "ERROR: timeout waiting for $name to reach status=$want (last status=$got)" >&2
      return 1
  fi
    sleep 5
  done
}

# Clean up old leftovers with the same names (optional, safe-ish)
if openstack server show "$COMPLIANT_NAME" >/dev/null 2>&1; then
  echo "Deleting existing $COMPLIANT_NAME ..."
  openstack server delete "$COMPLIANT_NAME" || true
fi
if openstack server show "$NONCOMPLIANT_NAME" >/dev/null 2>&1; then
  echo "Deleting existing $NONCOMPLIANT_NAME ..."
  openstack server delete "$NONCOMPLIANT_NAME" || true
fi

COMPLIANT_ID="$(create_server "$COMPLIANT_NAME")"
NONCOMPLIANT_ID="$(create_server "$NONCOMPLIANT_NAME")"

wait_status "$COMPLIANT_NAME" "ACTIVE"
wait_status "$NONCOMPLIANT_NAME" "ACTIVE"

echo "Stopping noncompliant server to make it SHUTOFF..."
openstack server stop "$NONCOMPLIANT_NAME"

# Wait until it is actually SHUTOFF
for _ in $(seq 1 60); do
  st="$(openstack server show "$NONCOMPLIANT_NAME" -f value -c status || true)"
  [[ "$st" == "SHUTOFF" ]] && break
  sleep 5
done

echo
echo "Created:"
echo "  compliant:    $COMPLIANT_NAME (id=$COMPLIANT_ID)"
echo "  noncompliant: $NONCOMPLIANT_NAME (id=$NONCOMPLIANT_ID, status=$(openstack server show "$NONCOMPLIANT_NAME" -f value -c status))"
echo
echo "Cleanup later with:"
echo "  openstack server delete $COMPLIANT_NAME $NONCOMPLIANT_NAME"