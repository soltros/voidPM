# voidPM (vpm)

A unified system helper and package management overlay for Void Linux.

voidPM unifies Runit service supervision, XBPS package management, Void Kernel branch control, xbps-src source package compilation, and system maintenance into a single command-line tool and terminal user interface.

## Features

- Runit Service Supervision (`vpm sv`): Status, enable, disable, start, stop, restart, reload, log tailing, user services.
- XBPS Package Management (`vpm pkg` / `vpm install`): Search, install, remove, system updates, metadata inspection, file ownership lookup, package holds.
- Kernel Management (`vpm kernel`): Inspect running/installed kernels, discover repository kernel branches, switch kernel series (`linux-lts`, `linux-mainline`), reconfigure initramfs/bootloader, regenerate dracut images, purge old kernels via `vkpurge`.
- Source Packages (`vpm src`): Setup `void-packages` repository, toggle `XBPS_ALLOW_RESTRICTED=yes`, build templates, install local `.xbps` binary packages.
- System Maintenance (`vpm clean`): Cache clearing, orphan removal, kernel purging, full system cleanup.
- Interactive TUI Dashboard (`vpm dashboard` / `vpm`): Terminal interface built with Bubble Tea and Lipgloss.

## Build and Installation

### Build Binary
```bash
make build
sudo make install
```

### Build Void Package (xbps-src)
```bash
vpm src setup
make package
vpm src install voidpm
```

## Command Reference

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
- `vpm kernel reconfigure [pkg]` - Reconfigure initramfs and bootloader hooks.
- `vpm kernel dracut` - Regenerate all dracut initramfs images.
- `vpm kernel purge [all|version]` - Remove old kernels via `vkpurge`.
- `vpm kernel hold` / `unhold` - Hold or unhold kernel metapackage.

### Source Packages (`vpm src`)
- `vpm src setup` - Initialize `void-packages` repository.
- `vpm src allow-restricted` - Enable restricted packages in `etc/conf`.
- `vpm src build <pkg> [-m]` - Build package from source.
- `vpm src install <pkg>` - Install compiled package from `hostdir/binpkgs`.
- `vpm src sync` - Update `void-packages` repository (`git pull`).

### System Maintenance (`vpm clean`)
- `vpm clean cache` - Remove obsolete package cache files (`xbps-remove -O`).
- `vpm clean orphans` - Remove unneeded orphan packages (`xbps-remove -o`).
- `vpm clean kernels` - Remove old kernels (`vkpurge rm all`).
- `vpm clean all` - Run all cleanup tasks.
