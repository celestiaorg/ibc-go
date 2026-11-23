#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."
test_name="${1:-}"
entrypoint="${2:-}"

command -v jq >/dev/null || { echo "jq is required to select e2e tests." >&2; exit 1; }
matrix=$(cd .. && go run -mod=readonly ./cmd/build_test_matrix)

if [[ "${RUN_SUITE:-}" == true ]]; then
  if [[ -z "$entrypoint" ]]; then
    command -v fzf >/dev/null || { echo "Specify entrypoint=TestSuiteName or install fzf." >&2; exit 1; }
    entrypoint=$(jq -r '.include[].entrypoint' <<< "$matrix" | sort -u | fzf)
  fi
  jq -e --arg suite "$entrypoint" '.include | any(.entrypoint == $suite)' <<< "$matrix" >/dev/null
else
  if [[ -z "$test_name" ]]; then
    command -v fzf >/dev/null || { echo "Specify test=TestName or install fzf." >&2; exit 1; }
    test_name=$(jq -r '.include[].test' <<< "$matrix" | sort -u | fzf)
  fi
  entrypoint=$(jq -er --arg test "$test_name" --arg suite "$entrypoint" '
    [.include[] | select(.test == $test and ($suite == "" or .entrypoint == $suite))]
    | if length == 1 then .[0].entrypoint else error("test must match exactly one suite") end
  ' <<< "$matrix")
fi

# Restrict discovery to test source files so logs/configs cannot be selected.
test_file=$(grep -rl --include='*_test.go' -F "func ${entrypoint}(" ./tests)
if [[ -z "$test_file" || "$test_file" == *$'\n'* ]]; then
  echo "Expected exactly one file for ${entrypoint}." >&2
  exit 1
fi
args=(-mod=readonly -tags=e2e -count=1 -v "$(dirname "$test_file")" -run "^${entrypoint}$" -timeout 30m)
if [[ "${RUN_SUITE:-}" != true ]]; then
  args+=(-testify.m "^${test_name}$")
fi

# Interchaintest v8 uses Docker SDK v24 (API 1.43). Newer daemons may
# require a higher minimum; preserve an explicit user override.
if [[ -z "${DOCKER_API_VERSION:-}" ]]; then
  minimum_api=$(docker version --format '{{.Server.MinAPIVersion}}')
  if [[ "$minimum_api" =~ ^([0-9]+)\.([0-9]+)$ ]] &&
    (( BASH_REMATCH[1] > 1 || BASH_REMATCH[2] > 43 )); then
    export DOCKER_API_VERSION="$minimum_api"
  fi
fi
go test "${args[@]}"
