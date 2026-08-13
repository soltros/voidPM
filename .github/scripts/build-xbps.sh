#!/bin/sh
set -ex

echo '==> Updating Void repositories and installing build dependencies...'
xbps-install -Sy
xbps-install -uy xbps
xbps-install -Sy bash git go make bsdtar coreutils

# Allow xbps-src to run with root permissions inside Docker container
export XBPS_ALLOW_CHROOT=yes

# Use ethereal (standard chroot) instead of uunshare to avoid /proc/self/uid_map writes
export XBPS_CHROOT_CMD=ethereal

BUILD_DIR="/tmp/void-packages"
rm -rf "${BUILD_DIR}"

echo '==> Cloning void-packages repository...'
git clone --depth=1 https://github.com/void-linux/void-packages.git "${BUILD_DIR}"

cd "${BUILD_DIR}"

# Append configuration settings to etc/conf
echo 'XBPS_ALLOW_CHROOT=yes' >> etc/conf
echo 'XBPS_CHROOT_CMD=ethereal' >> etc/conf

echo '==> Preparing srcpkgs/voidpm template...'
mkdir -p srcpkgs/voidpm/files
cp -a /workspace/template srcpkgs/voidpm/template
sed -i "s/^version=.*/version=${VERSION}/" srcpkgs/voidpm/template

echo '==> Copying voidPM source code...'
tar --exclude='./dist' --exclude='./.git' -cf - -C /workspace . | tar -xf - -C srcpkgs/voidpm/files/

echo '==> Bootstrapping xbps-src environment...'
./xbps-src binary-bootstrap

echo '==> Compiling voidpm XBPS package...'
./xbps-src pkg voidpm

echo '==> Copying built XBPS package artifacts to workspace dist...'
mkdir -p /workspace/dist
find hostdir/binpkgs -type f -name '*.xbps' -exec cp -v {} /workspace/dist/ \;
