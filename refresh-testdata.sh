#!/usr/bin/env bash
set -euo pipefail

# refresh-testdata.sh
#
# Regenerates plan.json (via Vertex AI Gemini) and expected.md (via local convert)
# for every test case under: testdata/convert-directives/<case>/
#
# Assumptions:
# - Google ADC credentials are available (gcloud auth application-default login OR GOOGLE_APPLICATION_CREDENTIALS)
# - Project is provided via env (GOOGLE_CLOUD_PROJECT or CLOUDSDK_CORE_PROJECT), or override with PROJECT env var
# - Uses Vertex AI location "global" and model "gemini-2.5-flash"
#
# Usage:
#   ./refresh-testdata.sh
#   PROJECT=my-project ./refresh-testdata.sh
#   ./refresh-testdata.sh 05-complex-doc 01-heading   # optionally limit to specific test case dirs (by name)

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTDATA_DIR="${ROOT_DIR}/testdata/convert-directives"

MODEL="${MODEL:-gemini-2.5-flash}"
LOCATION="${LOCATION:-global}"

PROJECT="${PROJECT:-${GOOGLE_CLOUD_PROJECT:-${CLOUDSDK_CORE_PROJECT:-}}}"
if [[ -z "${PROJECT}" ]]; then
  echo "ERROR: No GCP project set."
  echo "Set PROJECT, or GOOGLE_CLOUD_PROJECT, or CLOUDSDK_CORE_PROJECT."
  exit 2
fi

if [[ ! -d "${TESTDATA_DIR}" ]]; then
  echo "ERROR: Missing directory: ${TESTDATA_DIR}"
  exit 2
fi

if ! command -v go >/dev/null 2>&1; then
  echo "ERROR: go is not installed or not on PATH."
  exit 2
fi

echo "Project : ${PROJECT}"
echo "Location: ${LOCATION}"
echo "Model   : ${MODEL}"
echo

# If args are provided, treat them as case directory names to refresh.
# Otherwise refresh all subdirectories.
declare -a CASE_DIRS=()
if [[ $# -gt 0 ]]; then
  for name in "$@"; do
    CASE_DIRS+=("${TESTDATA_DIR}/${name}")
  done
else
  while IFS= read -r -d '' d; do
    CASE_DIRS+=("$d")
  done < <(find "${TESTDATA_DIR}" -mindepth 1 -maxdepth 1 -type d -print0 | sort -z)
fi

fail_count=0
for case_dir in "${CASE_DIRS[@]}"; do
  case_name="$(basename "${case_dir}")"

  echo Processing case: ${case_name}
  
  schema="${case_dir}/schema.json"
  instance="${case_dir}/instance.json"
  plan="${case_dir}/plan.json"
  expected="${case_dir}/expected.md"

  if [[ ! -f "${schema}" ]]; then
    echo "[SKIP] ${case_name}: missing schema.json"
    continue
  fi
  if [[ ! -f "${instance}" ]]; then
    echo "[SKIP] ${case_name}: missing instance.json"
    continue
  fi

  echo "==> Refreshing: ${case_name}"

  tmp_plan="$(mktemp)"
  tmp_expected="$(mktemp)"

  # 1) Generate a fresh plan.json by calling Gemini via Vertex AI (global region)
  (
    cd "${ROOT_DIR}"
    go run . --plan \
      --verbose \
      --schema-file "${schema}" \
      --pretty-print \
      --project "${PROJECT}" \
      --location "${LOCATION}" \
      --model "${MODEL}" \
      --out "${tmp_plan}"
  )

  # 2) Generate expected.md by running conversion locally (no LLM)
  (
    cd "${ROOT_DIR}"
    go run . --convert \
      --verbose \
      --json-file "${instance}" \
      --schema-file "${schema}" \
      --plan-file "${tmp_plan}" \
      --out "${tmp_expected}"
  )

  # 3) Atomically update golden files
  mv "${tmp_plan}" "${plan}"
  mv "${tmp_expected}" "${expected}"

  echo "    Updated: ${plan}"
  echo "    Updated: ${expected}"
  echo
done

echo "Done."