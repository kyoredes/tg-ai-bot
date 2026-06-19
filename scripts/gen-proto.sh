#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTOC="${ROOT}/.tools/bin/protoc"
export PATH="${ROOT}/.tools/bin:${HOME}/go/bin:${PATH}"

if [[ ! -x "${PROTOC}" ]]; then
  echo "protoc not found at ${PROTOC}. Download protoc to .tools/bin first."
  exit 1
fi

mkdir -p "${ROOT}/proto/gen/go/auth/v1"
mkdir -p "${ROOT}/proto/gen/go/subscription/v1"
mkdir -p "${ROOT}/proto/gen/go/ai/v1"

"${PROTOC}" \
  --proto_path="${ROOT}/proto" \
  --go_out="${ROOT}/proto/gen/go" --go_opt=module=rageai/proto/gen/go \
  --go-grpc_out="${ROOT}/proto/gen/go" --go-grpc_opt=module=rageai/proto/gen/go \
  "${ROOT}/proto/auth/v1/auth.proto" \
  "${ROOT}/proto/subscription/v1/subscription.proto" \
  "${ROOT}/proto/ai/v1/ai.proto"

echo "go proto generation complete"

if command -v python3 >/dev/null && python3 -m grpc_tools.protoc --help >/dev/null 2>&1; then
  bash "${ROOT}/ai-service/scripts/gen_proto.sh"
else
  echo "skip python proto generation (grpcio-tools not installed)"
fi
