#!/usr/bin/env sh
set -eu

case_id=$(printf '%s' "${1:-}" | tr -d '\r\n')
openclerk_cmd=${OCKP_PROMPTFOO_OPENCLERK_CMD:-openclerk}

run_openclerk() {
  domain=$1
  payload=$2
  printf '%s' "$payload" | sh -c "$openclerk_cmd $domain"
}

case "$case_id" in
  capabilities-authority-boundary)
    sh -c "$openclerk_cmd capabilities"
    ;;
  capabilities-promoted-workflows)
    sh -c "$openclerk_cmd capabilities"
    ;;
  document-validate-approved-create)
    run_openclerk document '{"action":"validate","document":{"path":"notes/promptfoo-smoke.md","title":"Promptfoo Smoke","body":"# Promptfoo Smoke"}}'
    ;;
  document-validate-missing-body)
    run_openclerk document '{"action":"validate","document":{"path":"notes/promptfoo-smoke.md","title":"Promptfoo Smoke"}}'
    ;;
  retrieval-validate-search)
    run_openclerk retrieval '{"action":"validate","search":{"text":"runner authority","limit":3}}'
    ;;
  retrieval-validate-negative-limit)
    run_openclerk retrieval '{"action":"validate","search":{"text":"runner authority","limit":-1}}'
    ;;
  public-source-validate-candidate)
    run_openclerk document '{"action":"validate","source":{"url":"https://example.com/openclerk-smoke.pdf","path_hint":"sources/promptfoo-smoke.md","source_type":"pdf","title":"Promptfoo Smoke Source"}}'
    ;;
  *)
    printf '{"rejected":true,"summary":"unknown promptfoo smoke case","case_id":"%s"}\n' "$case_id"
    exit 1
    ;;
esac
