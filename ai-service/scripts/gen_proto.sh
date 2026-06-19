#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_ROOT="$(cd "$ROOT/.." && pwd)/proto"
OUT_DIR="$ROOT/rpc"

mkdir -p "$OUT_DIR"

python3 -m grpc_tools.protoc \
  -I"$PROTO_ROOT" \
  --python_out="$OUT_DIR" \
  --grpc_python_out="$OUT_DIR" \
  "$PROTO_ROOT/ai/v1/ai.proto"

GRPC_FILE="$OUT_DIR/ai/v1/ai_pb2_grpc.py"
sed -i 's/from ai\.v1 import/from rpc.ai.v1 import/g' "$GRPC_FILE"

echo "python proto generation complete: $OUT_DIR/ai/v1"
