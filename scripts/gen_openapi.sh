#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if command -v oapi-codegen >/dev/null 2>&1; then
  GENERATOR=(oapi-codegen)
else
  GENERATOR=(go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest)
fi

shopt -s nullglob
specs=("$ROOT_DIR"/api/openapi/*/*.yaml)

if [ "${#specs[@]}" -eq 0 ]; then
  echo "no OpenAPI specs found under api/openapi/*/*.yaml"
  exit 0
fi

for spec in "${specs[@]}"; do
  node_name="$(basename "$(dirname "$spec")")"
  service_name="$(basename "$spec" .yaml)"
  package_name="${service_name}gen"
  output_dir="$ROOT_DIR/$node_name/internal/port/http/gen/$service_name"
  output_file="$output_dir/${service_name}_gen.go"

  mkdir -p "$output_dir"

  echo "generating $output_file from $spec"
  "${GENERATOR[@]}" \
    -package "$package_name" \
    -generate types,gin-server,spec \
    -o "$output_file" \
    "$spec"
done
