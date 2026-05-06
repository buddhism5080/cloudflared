#!/usr/bin/env bash
set -euo pipefail

TARGET="${1:-}"
OUTPUT_DIR="${2:-dist}"
ANDROID_API="${ANDROID_API:-21}"

if [[ -z "${TARGET}" ]]; then
  echo "usage: $0 <arm64|armv7> [output-dir]" >&2
  exit 64
fi

NDK_HOME="${ANDROID_NDK_HOME:-${ANDROID_NDK_ROOT:-}}"
if [[ -n "${NDK_HOME}" ]]; then
  NDK_BIN="${NDK_HOME}/toolchains/llvm/prebuilt/linux-x86_64/bin"
else
  NDK_BIN=""
fi

TARGET_OS="android"
TARGET_ARCH=""
TARGET_ARM=""
DEFAULT_CGO_ENABLED="0"
ARTIFACT_NAME=""

case "${TARGET}" in
  arm64)
    TARGET_ARCH="arm64"
    ARTIFACT_NAME="cloudflared-android-arm64"
    ;;
  armv7)
    TARGET_ARCH="arm"
    TARGET_ARM="7"
    DEFAULT_CGO_ENABLED="1"
    ARTIFACT_NAME="cloudflared-android-armv7"
    if [[ -z "${NDK_BIN}" ]]; then
      echo "ANDROID_NDK_HOME or ANDROID_NDK_ROOT is required for ${TARGET}" >&2
      exit 1
    fi
    export CC="${CC:-${NDK_BIN}/armv7a-linux-androideabi${ANDROID_API}-clang}"
    export CXX="${CXX:-${NDK_BIN}/armv7a-linux-androideabi${ANDROID_API}-clang++}"
    ;;
  *)
    echo "unsupported target: ${TARGET}" >&2
    exit 64
    ;;
esac

export CGO_ENABLED="${CGO_ENABLED:-${DEFAULT_CGO_ENABLED}}"
export TARGET_OS
export TARGET_ARCH
if [[ -n "${TARGET_ARM}" ]]; then
  export TARGET_ARM
fi

if [[ -z "${STRIP_BIN:-}" && -n "${NDK_BIN}" ]]; then
  STRIP_BIN="${NDK_BIN}/llvm-strip"
fi

mkdir -p "${OUTPUT_DIR}"
rm -f cloudflared "${OUTPUT_DIR}/${ARTIFACT_NAME}" "${OUTPUT_DIR}/${ARTIFACT_NAME}.sha256"

echo "Building ${TARGET_OS}/${TARGET_ARCH}${TARGET_ARM:+ GOARM=${TARGET_ARM}} with CGO_ENABLED=${CGO_ENABLED}"
if [[ -n "${CC:-}" ]]; then
  echo "CC=${CC}"
fi

make cloudflared
mv cloudflared "${OUTPUT_DIR}/${ARTIFACT_NAME}"

if [[ -n "${STRIP_BIN:-}" && -x "${STRIP_BIN}" ]]; then
  "${STRIP_BIN}" "${OUTPUT_DIR}/${ARTIFACT_NAME}" || true
fi

sha256sum "${OUTPUT_DIR}/${ARTIFACT_NAME}" | tee "${OUTPUT_DIR}/${ARTIFACT_NAME}.sha256"

if command -v readelf >/dev/null 2>&1; then
  readelf -h "${OUTPUT_DIR}/${ARTIFACT_NAME}" | sed -n '1,20p'
fi
