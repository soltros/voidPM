# VPM (Void Poweruser Management System)

A unified system helper, kernel series manager, runit supervisor, and package management overlay for Void Linux.

**VPM** (`vpm`) unifies Runit service supervision, XBPS package management, Void Kernel series branch control, `xbps-src` source package compilation, and system maintenance into a single command-line tool and interactive terminal user interface (TUI).

## Features

- Runit Service Supervision (`vpm sv`): Status, enable, disable, start, stop, restart, reload, log tailing, user services.
- XBPS Package Management (`vpm pkg` / `vpm install`): Search, install, remove, system updates, metadata inspection, file ownership lookup, package holds.
- Kernel Management (`vpm kernel`): Inspect running/installed kernels, discover repository kernel branches, switch kernel series (`linux-lts`, `linux-mainline`), reconfigure initramfs/bootloader, regenerate dracut images, purge old kernels via `vkpurge`.
- Source Packages (`vpm src`): Setup `void-packages` repository, toggle `XBPS_ALLOW_RESTRICTED=yes`, build templates, install local `.xbps` binary packages.
- System Maintenance (`vpm clean`): Cache clearing, orphan removal, kernel purging, full system cleanup.
- Self-Update (`vpm self-update`): Automatic GitHub API release fetching & binary overwrite into `/usr/bin/vpm`.
- Interactive TUI Dashboard (`vpm dashboard` / `vpm`): Terminal interface built with Bubble Tea and Lipgloss.

## Installation

### Option 1: Install XBPS Package (Recommended)
Download and install the pre-compiled `.xbps` binary package directly:
```bash
# Download the latest release package
curl -LO https://github.com/soltros/voidPM/releases/latest/download/voidpm-0.1.0_1.x86_64.xbps

# Install via xbps-install
sudo xbps-install -y ./voidpm-0.1.0_1.x86_64.xbps
```

### Option 2: Direct Binary Installation
Install the compiled `vpm` executable directly into `/usr/bin`:
```bash
sudo curl -L -o /usr/bin/vpm https://github.com/soltros/voidPM/releases/latest/download/vpm
sudo chmod +x /usr/bin/vpm
```

### Option 3: Self-Update
If `vpm` is already installed on your system, perform a self-update at any time:
```bash
vpm self-update
```

### Option 4: Build from Source
```bash
# Build and install local binary:
make build
sudo make install

# Or build local XBPS package using xbps-src:
make package
sudo xbps-install -y --repository=dist voidpm
```


## Command Reference

### Self-Update (`vpm self-update`)
- `vpm self-update` - Query GitHub API (`soltros/voidPM`), download latest release binary, and overwrite `/usr/bin/vpm`.
- `vpm sync --self` / `vpm update -s` - Perform self-update before system upgrade.

### Runit Services (`vpm sv`)
- `vpm sv status` - Show status table of all runit services.
- `vpm sv enable <service>` - Enable service (`/etc/sv/<name>` -> `/var/service/`).
- `vpm sv disable <service>` - Disable service (remove `/var/service/<name>`).
- `vpm sv start <service>` - Start service (`sv up`).
- `vpm sv stop <service>` - Stop service (`sv down`).
- `vpm sv restart <service>` - Restart service (`sv restart`).
- `vpm sv reload <service>` - Reload service config (`sv reload`).
- `vpm sv log <service>` - Tail service logs.

### Package Management (`vpm pkg` / `vpm install` / `vpm remove`)
- `vpm search <query>` - Search repository packages.
- `vpm install <package...>` - Install or update packages.
- `vpm remove <package...>` - Remove packages (`-R` for recursive removal).
- `vpm update` - Synchronize repositories and upgrade system (`xbps-install -Su`).
- `vpm info <package>` - Show package metadata.
- `vpm orphans` - List orphaned packages.
- `vpm whoowns <path>` - Find package owning a file.
- `vpm files <package>` - List files in package.
- `vpm hold <package>` - Hold package updates.
- `vpm unhold <package>` - Remove package hold.

### Kernel Management (`vpm kernel`)
- `vpm kernel status` - Show running and installed kernels.
- `vpm kernel available` - List available kernel series in repositories.
- `vpm kernel switch <series>` - Switch active kernel series (e.g., `linux-lts`).
- `vpm kernel remove <series>` - Safely uninstall a specific kernel package/series (`vpm kernel remove linux7.1`).
- `vpm kernel reconfigure [pkg]` - Reconfigure initramfs and bootloader hooks.
- `vpm kernel dracut` - Regenerate all dracut initramfs images.
- `vpm kernel purge [all|version]` - Remove old kernels via `vkpurge`.
- `vpm kernel hold` / `unhold` - Hold or unhold kernel metapackage.

### Void Source Packages (`vpm src`)
Manages non-distributable and custom source builds via `void-packages` (stored cleanly in `~/.void-packages`):
- `vpm src setup` - Initialize `void-packages` repo into `~/.void-packages` and run binary bootstrap (auto-runs when needed).
- `vpm src search <query>` - Rich template search in `srcpkgs/` with status, restricted badges, and descriptions.
- `vpm src build <pkg>` - Build package from source (automatically detects `restricted=yes` licenses and passes required flags).
- `vpm src install <pkg>` - Install built binary package from local `hostdir/binpkgs`.
- `vpm src allow-restricted` - Enable restricted packages in `etc/conf`.
- `vpm src sync` - Update source package templates via `git pull`.

### System Maintenance (`vpm clean`)
- `vpm clean cache` - Remove obsolete package cache files (`xbps-remove -O`).
- `vpm clean orphans` - Remove unneeded orphan packages (`xbps-remove -o`).
- `vpm clean kernels` - Remove old kernels (`vkpurge rm all`).
- `vpm clean all` - Run all cleanup tasks.

## License

Distributed under the GNU General Public License v3.0 (GPL-3.0-or-later). See `LICENSE` for details.
