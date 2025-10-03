#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: build_and_push_images.sh --repo <dockerhub-user> --version <tag> [--latest]

Prerequisites:
  - docker login (once per session)
  - Oracle Instant Client ZIP at orc_client/instantclient-basic-linux.x64-*.zip

Flags:
  --repo <value>     Docker Hub namespace (e.g. myuser)
  --version <value>  Image tag to publish (e.g. 2025.03.14)
  --latest           Also tag and push :latest alongside the provided version
USAGE
}

repo=""
version=""
push_latest="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo)
      repo=${2:-}
      shift 2
      ;;
    --version)
      version=${2:-}
      shift 2
      ;;
    --latest)
      push_latest="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "$repo" || -z "$version" ]]; then
  echo "--repo and --version are required." >&2
  usage
  exit 1
fi

if ! compgen -G "orc_client/instantclient-basic-linux.x64-*.zip" > /dev/null; then
  echo "Missing Oracle Instant Client ZIP under orc_client/." >&2
  exit 1
fi

api_image="${repo}/bigmeter-api:${version}"
sync_image="${repo}/bigmeter-sync-thick:${version}"

build_image() {
  local dockerfile=$1
  local image=$2
  echo "[build] ${image}"
  docker build -f "${dockerfile}" -t "${image}" .
}

push_image() {
  local image=$1
  echo "[push] ${image}"
  docker push "${image}"
}

build_image docker/Dockerfile.api "${api_image}"
build_image docker/Dockerfile.sync-thick "${sync_image}"

push_image "${api_image}"
push_image "${sync_image}"

if [[ "${push_latest}" == "true" ]]; then
  for suffix in api sync-thick; do
    image="${repo}/bigmeter-${suffix}"
    latest="${image}:latest"
    versioned="${image}:${version}"
    echo "[tag] ${latest}"
    docker tag "${versioned}" "${latest}"
    push_image "${latest}"
  done
fi

echo "Done. Published:"
echo "  ${api_image}"
echo "  ${sync_image}"
if [[ "${push_latest}" == "true" ]]; then
  echo "  ${repo}/bigmeter-api:latest"
  echo "  ${repo}/bigmeter-sync-thick:latest"
fi
