#!/usr/bin/env bash
set -euo pipefail

#
# Creates a small set of OpenStack resources for OSPA e2e tests, based on the
# currently-covered services/resources in this repo.
#
# What it does:
#  1) Detects covered services/resources from the codebase (generated/implemented).
#  2) Creates a couple of real OpenStack resources using `openstack` CLI so e2e tests
#     can discover/audit them.
#
# Notes:
# - This script is best-effort and intentionally conservative. It only automates resources
#   that have well-known `openstack` CLI create commands.
# - For other resources in the scaffold registry, it will print a "not automated yet" message.
#
# Requirements:
# - `openstack` CLI installed and configured (clouds.yaml + OS_CLOUD)
# - Quota/permissions to create resources in the target project
#
# Usage:
#   export OS_CLOUD=mycloud
#   bash ./scripts/create_test_cluster_based_on_coverage.sh
#
# Optional inputs:
#   export NAME_PREFIX=ospa-e2e
#   export IMAGE_NAME_OR_ID=...
#   export FLAVOR_NAME_OR_ID=...
#   export NETWORK_NAME_OR_ID=...
#   export EXT_NETWORK_NAME_OR_ID=...   # external network for floating IP allocation
#   export VOLUME_SIZE=1                # GiB
#   export KEY_NAME=...                 # optional (used for servers)
#   export SECURITY_GROUP=...           # optional (used for servers)
#   export TIMEOUT_SECS=600
#   export FILTER_SERVICE=nova          # optional: limit to a single service
#   export FILTER_RESOURCE=instance     # optional: limit to a single resource
#   export SKIP_CREATE=1                # optional: only print coverage list

: "${OS_CLOUD:?OS_CLOUD must be set (cloud name in clouds.yaml)}"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

NAME_PREFIX="${NAME_PREFIX:-ospa-e2e}"
VOLUME_SIZE="${VOLUME_SIZE:-1}"

if ! command -v openstack >/dev/null 2>&1; then
  echo "ERROR: openstack CLI not found in PATH" >&2
  exit 1
fi

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

pick_external_network() {
  local ext="${EXT_NETWORK_NAME_OR_ID:-}"
  if [[ -n "$ext" ]]; then
    echo "$ext"
    return 0
  fi
  ext="$(bash -lc "openstack network list --external -f value -c ID" | head -n1 || true)"
  if [[ -z "$ext" ]]; then
    echo "ERROR: Could not auto-pick EXT_NETWORK_NAME_OR_ID (no external networks found)." >&2
    return 1
  fi
  echo "$ext"
}

detect_covered_resources() {
  # Covered = resources that have an auditor file (primary signal),
  # optionally filtered by service/resource.
  local covered=()
  local svc
  for svcdir in "${ROOT}/pkg/audit/"*; do
    [[ -d "${svcdir}" ]] || continue
    svc="$(basename "${svcdir}")"
    local f
    for f in "${svcdir}/"*.go; do
      [[ -f "${f}" ]] || continue
      [[ "${f}" == *"_test.go" ]] && continue
      local res
      res="$(basename "${f}" .go)"
      if [[ -n "${FILTER_SERVICE:-}" && "${FILTER_SERVICE}" != "${svc}" ]]; then
        continue
      fi
      if [[ -n "${FILTER_RESOURCE:-}" && "${FILTER_RESOURCE}" != "${res}" ]]; then
        continue
      fi
      covered+=("${svc}:${res}")
    done
  done

  # Print one per line for easy consumption
  printf "%s\n" "${covered[@]:-}"
}

echo "==> Coverage detection (from pkg/audit/*/*.go)"
covered_list="$(detect_covered_resources || true)"
if [[ -z "${covered_list}" ]]; then
  echo "No covered resources detected."
else
  echo "${covered_list}" | sed 's/^/  - /'
fi
echo

if [[ "${SKIP_CREATE:-0}" == "1" ]]; then
  echo "SKIP_CREATE=1 set; not creating any resources."
  exit 0
fi

echo "==> Creating resources for detected coverage (best-effort)"
echo "Using:"
echo "  OS_CLOUD=$OS_CLOUD"
echo "  NAME_PREFIX=$NAME_PREFIX"
echo "  VOLUME_SIZE=$VOLUME_SIZE"
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

create_keypair() {
  local name="$1"
  echo "Creating keypair: $name"
  if openstack keypair show "$name" >/dev/null 2>&1; then
    echo "Deleting existing keypair $name ..."
    openstack keypair delete "$name" || true
  fi
  openstack keypair create "$name" -f value -c name >/dev/null
}

create_security_group() {
  local name="$1"
  echo "Creating security group: $name"
  local id
  id="$(openstack security group create "$name" -f value -c id 2>/dev/null || true)"
  if [[ -z "$id" ]]; then
    echo "ERROR: failed to create security group $name" >&2
    return 1
  fi
  echo "$id"
}

create_security_group_rule() {
  local sg_id="$1"
  local proto="$2"
  local port="$3"
  local cidr="$4"
  # openstack security group rule create [--ingress|--egress] --protocol tcp --dst-port 22 --remote-ip 0.0.0.0/0 <sg>
  openstack security group rule create --ingress --protocol "$proto" --dst-port "$port" --remote-ip "$cidr" "$sg_id" -f value -c id >/dev/null
}

create_floating_ip() {
  local ext_net="$1"
  local id
  id="$(openstack floating ip create "$ext_net" -f value -c id 2>/dev/null || true)"
  if [[ -z "$id" ]]; then
    echo "ERROR: failed to allocate floating IP on external network $ext_net" >&2
    return 1
  fi
  echo "$id"
}

wait_volume_status() {
  local id="$1"
  local want="$2"
  local timeout_secs="${TIMEOUT_SECS:-600}"
  local start end got
  start="$(date +%s)"
  end="$((start + timeout_secs))"
  while true; do
    got="$(openstack volume show "$id" -f value -c status 2>/dev/null || true)"
    if [[ "$got" == "$want" ]]; then
      return 0
    fi
    if [[ "$(date +%s)" -ge "$end" ]]; then
      echo "ERROR: timeout waiting for volume $id to reach status=$want (last status=$got)" >&2
      return 1
    fi
    sleep 5
  done
}

wait_snapshot_status() {
  local id="$1"
  local want="$2"
  local timeout_secs="${TIMEOUT_SECS:-600}"
  local start end got
  start="$(date +%s)"
  end="$((start + timeout_secs))"
  while true; do
    got="$(openstack volume snapshot show "$id" -f value -c status 2>/dev/null || true)"
    if [[ "$got" == "$want" ]]; then
      return 0
    fi
    if [[ "$(date +%s)" -ge "$end" ]]; then
      echo "ERROR: timeout waiting for snapshot $id to reach status=$want (last status=$got)" >&2
      return 1
    fi
    sleep 5
  done
}

create_volume() {
  local name="$1"
  local size="$2"
  echo "Creating volume: $name (size=${size}GiB)"
  local id
  id="$(openstack volume create --size "$size" "$name" -f value -c id 2>/dev/null || true)"
  if [[ -z "$id" ]]; then
    echo "ERROR: failed to create volume $name" >&2
    return 1
  fi
  echo "$id"
}

create_snapshot() {
  local volume_id="$1"
  local name="$2"
  echo "Creating snapshot: $name (volume=$volume_id)"
  local id
  id="$(openstack volume snapshot create --volume "$volume_id" "$name" -f value -c id 2>/dev/null || true)"
  if [[ -z "$id" ]]; then
    echo "ERROR: failed to create snapshot $name" >&2
    return 1
  fi
  echo "$id"
}

# Resolve image/flavor/network only if we need to create servers
need_servers=0
if echo "${covered_list}" | grep -q "^nova:instance$"; then
  need_servers=1
fi

IMAGE="${IMAGE_NAME_OR_ID:-}"
FLAVOR="${FLAVOR_NAME_OR_ID:-}"
NETWORK="${NETWORK_NAME_OR_ID:-}"
if [[ "$need_servers" == "1" ]]; then
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
  echo "  IMAGE=$IMAGE"
  echo "  FLAVOR=$FLAVOR"
  echo "  NETWORK=$NETWORK"
  echo
fi

# Track created IDs/names for summary
CREATED=()

# Create in a stable order
while IFS= read -r item; do
  [[ -n "$item" ]] || continue
  svc="${item%%:*}"
  res="${item##*:}"

  case "${svc}:${res}" in
    nova:instance)
      COMPLIANT_NAME="${NAME_PREFIX}-nova-instance-compliant"
      NONCOMPLIANT_NAME="${NAME_PREFIX}-nova-instance-noncompliant"

      # Clean up old leftovers by name
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

      # Make one SHUTOFF to support status-based policies later.
      echo "Stopping noncompliant server to make it SHUTOFF..."
      openstack server stop "$NONCOMPLIANT_NAME" || true

      CREATED+=("server:${COMPLIANT_NAME}:${COMPLIANT_ID}")
      CREATED+=("server:${NONCOMPLIANT_NAME}:${NONCOMPLIANT_ID}")
      ;;

    nova:keypair)
      kp1="${NAME_PREFIX}-nova-keypair-1"
      kp2="${NAME_PREFIX}-nova-keypair-2"
      create_keypair "$kp1"
      create_keypair "$kp2"
      CREATED+=("keypair:${kp1}")
      CREATED+=("keypair:${kp2}")
      ;;

    neutron:security_group)
      sg_name="${NAME_PREFIX}-neutron-sg"
      # Delete existing by name if present
      if openstack security group show "$sg_name" >/dev/null 2>&1; then
        echo "Deleting existing security group $sg_name ..."
        openstack security group delete "$sg_name" || true
      fi
      sg_id="$(create_security_group "$sg_name")"
      CREATED+=("security_group:${sg_name}:${sg_id}")
      ;;

    neutron:security_group_rule)
      # Create SG if not already created
      sg_name="${NAME_PREFIX}-neutron-sg"
      if openstack security group show "$sg_name" >/dev/null 2>&1; then
        sg_id="$(openstack security group show "$sg_name" -f value -c id)"
      else
        sg_id="$(create_security_group "$sg_name")"
        CREATED+=("security_group:${sg_name}:${sg_id}")
      fi
      # Add two simple ingress rules
      create_security_group_rule "$sg_id" tcp 22 "0.0.0.0/0"
      create_security_group_rule "$sg_id" tcp 80 "0.0.0.0/0"
      CREATED+=("security_group_rule:sg=${sg_id}:(tcp/22, tcp/80)")
      ;;

    neutron:floating_ip)
      ext_net="$(pick_external_network)"
      fip1="$(create_floating_ip "$ext_net")"
      fip2="$(create_floating_ip "$ext_net")"
      CREATED+=("floating_ip:${fip1}")
      CREATED+=("floating_ip:${fip2}")
      ;;

    cinder:volume)
      v1="${NAME_PREFIX}-cinder-volume-1"
      v2="${NAME_PREFIX}-cinder-volume-2"
      # Delete existing by name (best-effort)
      if openstack volume show "$v1" >/dev/null 2>&1; then openstack volume delete "$v1" || true; fi
      if openstack volume show "$v2" >/dev/null 2>&1; then openstack volume delete "$v2" || true; fi
      vol1_id="$(create_volume "$v1" "$VOLUME_SIZE")"
      vol2_id="$(create_volume "$v2" "$VOLUME_SIZE")"
      wait_volume_status "$vol1_id" "available"
      wait_volume_status "$vol2_id" "available"
      CREATED+=("volume:${v1}:${vol1_id}")
      CREATED+=("volume:${v2}:${vol2_id}")
      ;;

    cinder:snapshot)
      # Ensure at least one volume exists
      v1="${NAME_PREFIX}-cinder-volume-1"
      if openstack volume show "$v1" >/dev/null 2>&1; then
        vol1_id="$(openstack volume show "$v1" -f value -c id)"
      else
        vol1_id="$(create_volume "$v1" "$VOLUME_SIZE")"
        wait_volume_status "$vol1_id" "available"
        CREATED+=("volume:${v1}:${vol1_id}")
      fi
      snap_name="${NAME_PREFIX}-cinder-snapshot-1"
      if openstack volume snapshot show "$snap_name" >/dev/null 2>&1; then
        echo "Deleting existing snapshot $snap_name ..."
        openstack volume snapshot delete "$snap_name" || true
      fi
      snap_id="$(create_snapshot "$vol1_id" "$snap_name")"
      wait_snapshot_status "$snap_id" "available"
      CREATED+=("snapshot:${snap_name}:${snap_id}")
      ;;

    *)
      echo "NOTE: Not automated yet for ${svc}:${res}. (Known coverage, but no create logic in this script.)"
      ;;
  esac
done <<< "${covered_list}"

echo
echo "==> Created resources:"
for item in "${CREATED[@]:-}"; do
  echo "  - $item"
done

echo
echo "==> Cleanup hints (manual):"
echo "  # servers"
echo "  openstack server delete ${NAME_PREFIX}-nova-instance-compliant ${NAME_PREFIX}-nova-instance-noncompliant"
echo "  # keypairs"
echo "  openstack keypair delete ${NAME_PREFIX}-nova-keypair-1 ${NAME_PREFIX}-nova-keypair-2"
echo "  # neutron"
echo "  openstack security group delete ${NAME_PREFIX}-neutron-sg"
echo "  # floating IPs: list and delete those matching prefix (no tag in this script)"
echo "  openstack floating ip list"
echo "  # cinder"
echo "  openstack volume snapshot delete ${NAME_PREFIX}-cinder-snapshot-1"
echo "  openstack volume delete ${NAME_PREFIX}-cinder-volume-1 ${NAME_PREFIX}-cinder-volume-2"
