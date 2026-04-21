#!/usr/bin/env bash
set -eEuo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <service-name> <output-dir>" >&2
  exit 1
fi

SERVICE_NAME="$1"
OUT_DIR="$2"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd -P)"
BLUEPRINT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd -P)"
SRC="${BLUEPRINT_DIR}/templates/service/sample-service"

if [[ -e "${OUT_DIR}" ]]; then
  echo "output already exists: ${OUT_DIR}" >&2
  exit 1
fi

PACKAGE_NAME="$(echo "${SERVICE_NAME}" | sed -E 's/[^a-zA-Z0-9]+//g')"
if [[ -z "${PACKAGE_NAME}" ]]; then
  echo "service name must contain at least one alphanumeric character" >&2
  exit 1
fi

cp -R "${SRC}" "${OUT_DIR}"

find "${OUT_DIR}" -type f -print0 | xargs -0 perl -pi -e "s/sample-service/${SERVICE_NAME}/g; s/package sample/package ${PACKAGE_NAME}/g; s/internal\\/sample/internal\\/${PACKAGE_NAME}/g; s/sample\\.NewServer/${PACKAGE_NAME}.NewServer/g; s/sample\\.Config/${PACKAGE_NAME}.Config/g"

if [[ -d "${OUT_DIR}/internal/sample" ]]; then
  mv "${OUT_DIR}/internal/sample" "${OUT_DIR}/internal/${PACKAGE_NAME}"
fi
if [[ -d "${OUT_DIR}/cmd/sample-service" ]]; then
  mv "${OUT_DIR}/cmd/sample-service" "${OUT_DIR}/cmd/${SERVICE_NAME}"
fi

echo "created ${SERVICE_NAME} at ${OUT_DIR}"
echo "next: cd ${OUT_DIR} && go test ./..."
