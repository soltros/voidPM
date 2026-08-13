#!/usr/bin/env bash
set -e

PKGNAME="voidpm"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Detect version and revision from template
VERSION="$(grep '^version=' "${ROOT_DIR}/template" | cut -d'=' -f2)"
REVISION="$(grep '^revision=' "${ROOT_DIR}/template" | cut -d'=' -f2)"

# Determine void-packages directory location
VOID_PKGS_DIR="${VOIDPM_SRC_DIR:-$HOME/.void-packages}"
if [ ! -d "${VOID_PKGS_DIR}" ] && [ -d "$HOME/void-packages" ]; then
    VOID_PKGS_DIR="$HOME/void-packages"
fi

echo "=================================================="
echo " Packaging voidPM v${VERSION} (rev ${REVISION}) for Void Linux"
echo " Target void-packages: ${VOID_PKGS_DIR}"
echo "=================================================="

# Ensure void-packages repo is initialized
if [ ! -f "${VOID_PKGS_DIR}/xbps-src" ]; then
    echo "--> Initializing void-packages repository in ${VOID_PKGS_DIR}..."
    mkdir -p "$(dirname "${VOID_PKGS_DIR}")"
    git clone --depth=1 https://github.com/void-linux/void-packages.git "${VOID_PKGS_DIR}"
    echo "--> Running binary-bootstrap..."
    (cd "${VOID_PKGS_DIR}" && ./xbps-src binary-bootstrap)
fi

TARGET_SRC_DIR="${VOID_PKGS_DIR}/srcpkgs/${PKGNAME}"
mkdir -p "${TARGET_SRC_DIR}/files"

echo "--> Copying package template..."
cp -a "${ROOT_DIR}/template" "${TARGET_SRC_DIR}/template"

echo "--> Copying source files..."
tar --exclude='./dist' --exclude='./.git' -cf - -C "${ROOT_DIR}" . | tar -xf - -C "${TARGET_SRC_DIR}/files/"

echo "--> Building XBPS package using xbps-src..."
(cd "${VOID_PKGS_DIR}" && ./xbps-src pkg "${PKGNAME}")

OUTPUT_DIR="${ROOT_DIR}/dist"
mkdir -p "${OUTPUT_DIR}"

echo "--> Collecting built XBPS package..."
find "${VOID_PKGS_DIR}/hostdir/binpkgs" -type f -name "${PKGNAME}-*.xbps" -exec cp -v {} "${OUTPUT_DIR}/" \;

# Optional publishing to GitHub Releases via gh CLI
if [ "$1" = "--publish" ] || [ "$1" = "-p" ]; then
    TAG="${2:-v${VERSION}}"
    echo "--> Building Go binary for release..."
    go build -ldflags="-s -w" -o "${OUTPUT_DIR}/vpm" "${ROOT_DIR}"
    (cd "${OUTPUT_DIR}" && tar -czvf "vpm-linux-amd64.tar.gz" vpm && sha256sum * > SHA256SUMS)

    echo "--> Publishing release ${TAG} to GitHub Releases..."
    if gh release view "${TAG}" &>/dev/null; then
        (cd "${OUTPUT_DIR}" && gh release upload "${TAG}" ./* --clobber)
    else
        (cd "${OUTPUT_DIR}" && gh release create "${TAG}" ./* --title "voidPM ${TAG}" --generate-notes)
    fi
fi

echo "=================================================="
echo " Package build complete!"
echo " Artifacts saved in: ${OUTPUT_DIR}"
echo "=================================================="
ls -lh "${OUTPUT_DIR}"
