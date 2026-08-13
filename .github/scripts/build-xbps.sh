#!/bin/sh
set -ex

echo '==> Updating Void repositories and installing build dependencies...'
xbps-install -Sy
xbps-install -uy xbps
xbps-install -Sy bash git go make bsdtar coreutils

# Create a non-root user to run xbps-src
useradd -m -s /bin/bash builder || true

BUILD_DIR="/tmp/void-packages"
rm -rf "${BUILD_DIR}"

echo '==> Cloning void-packages repository...'
git clone --depth=1 https://github.com/void-linux/void-packages.git "${BUILD_DIR}"

echo '==> Preparing srcpkgs/voidpm template...'
mkdir -p "${BUILD_DIR}/srcpkgs/voidpm/files"
cp -a /workspace/template "${BUILD_DIR}/srcpkgs/voidpm/template"
sed -i "s/^version=.*/version=${VERSION}/" "${BUILD_DIR}/srcpkgs/voidpm/template"

echo '==> Copying voidPM source code...'
tar --exclude='./dist' --exclude='./.git' -cf - -C /workspace . | tar -xf - -C "${BUILD_DIR}/srcpkgs/voidpm/files/"

# Change ownership of build directories to the non-root user
chown -R builder:builder "${BUILD_DIR}" /workspace

echo '==> Bootstrapping xbps-src environment...'
cd "${BUILD_DIR}"
su - builder -c "cd ${BUILD_DIR} && ./xbps-src binary-bootstrap"

echo '==> Compiling voidpm XBPS package...'
su - builder -c "cd ${BUILD_DIR} && ./xbps-src pkg voidpm"

echo '==> Copying built XBPS package artifacts to workspace dist...'
mkdir -p /workspace/dist
find "${BUILD_DIR}/hostdir/binpkgs" -type f -name '*.xbps' -exec cp -v {} /workspace/dist/ \;
