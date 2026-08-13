#!/bin/sh
set -ex

echo '==> Updating Void repositories and installing build dependencies...'
xbps-install -Sy
xbps-install -uy xbps
xbps-install -Sy bash git go make bsdtar coreutils shadow

echo '==> Creating unprivileged builder user...'
if ! id -u builder >/dev/null 2>&1; then
    useradd -m -s /bin/bash builder
fi

BUILD_DIR="/home/builder/void-packages"
rm -rf "${BUILD_DIR}"

echo '==> Cloning void-packages repository...'
git clone --depth=1 https://github.com/void-linux/void-packages.git "${BUILD_DIR}"

echo '==> Preparing srcpkgs/voidpm template...'
mkdir -p "${BUILD_DIR}/srcpkgs/voidpm/files"
cp -a /workspace/template "${BUILD_DIR}/srcpkgs/voidpm/template"
sed -i "s/^version=.*/version=${VERSION}/" "${BUILD_DIR}/srcpkgs/voidpm/template"

echo '==> Copying voidPM source code...'
tar --exclude='./dist' --exclude='./.git' -cf - -C /workspace . | tar -xf - -C "${BUILD_DIR}/srcpkgs/voidpm/files/"

echo '==> Setting ownership for builder user...'
chown -R builder:builder /home/builder

echo '==> Compiling voidpm XBPS package (non-chroot mode)...'
su - builder -c "cd ${BUILD_DIR} && ./xbps-src -N pkg voidpm"

echo '==> Copying built XBPS package artifacts to workspace dist...'
mkdir -p /workspace/dist
find "${BUILD_DIR}/hostdir/binpkgs" -type f -name '*.xbps' -exec cp -v {} /workspace/dist/ \;
