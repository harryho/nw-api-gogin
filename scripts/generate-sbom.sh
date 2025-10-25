#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$ROOT_DIR"

CYCLO_CMD=$(command -v cyclonedx-gomod || true)
if [[ -z "$CYCLO_CMD" ]]; then
  go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.6.0
  CYCLO_CMD="$(go env GOPATH)/bin/cyclonedx-gomod"
fi

OUTPUT_DIR=${OUTPUT_DIR:-sbom}
mkdir -p "$OUTPUT_DIR"

"$CYCLO_CMD" mod \
  -licenses \
  -json \
  -output "$OUTPUT_DIR/bom.json"

echo "SBOM written to $OUTPUT_DIR/bom.json"
