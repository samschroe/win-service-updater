# win-service-updater

![](https://github.com/huntresslabs/win-service-updater/workflows/Build/badge.svg)
![](https://github.com/huntresslabs/win-service-updater/workflows/Test/badge.svg)

Implementation of "core" wyUpdate functionality. This updater is written in GoLang to avoid .NET dependencies.

## Goals

- Support for TLS1.3 (default with GoLang 1.13 and main reason for creating this updater)
- Drop in replacement for use in existing service update commands (should work for any update, there is just no GUI)
- Compatibility with existing wyUpdate binary files

## Features Implemented

Basic update functionality works.

- Check for update only (`/justcheck /quickcheck` arguments)
- Replacement of `%urlargs%` in URLs when `-urlargs` argument is provided
- Update file signature verification
- Full file update with ability to stop/start services before/after the update
- Rollback on failure
- Logging (`-logging` and `/outputinfo` arguments)

## Current Limitations/Differences

- Support only for A.B.C.D (or A.B.C) version numbering; No "alpha", "beta", "pre", etc.
- No GUI component, created to be run from a service or command-line
- Only full binary replacement (no diff)
- Only supports stopping/starting services before/after update
  - No registry updates, COM updates, etc.
- No functionality to elevate privileges (need admin to install a service anyway)
- Only "installs" files to the "base directory" (directory the updater is run from)
- No FTP support
- Does not self-update
- No uninstall

## Arguments supported

- "/quickcheck"
- "/justcheck"
- "/noerr",
- "-urlargs=_args_"
- "/outputinfo=_out_"
- "/fromservice" (normal operation, but added so the argument parser doesn't error)
- "-logfile=_log_"
- "-cdata=_file_"
- "-server=_url_"

## General Operation

- To check if an update is available:
  - Download the .wys file (update URL specified in client.wyc)
  - Compare the available update (speficied in the .wys file) with the version currently installed
- If an update is required:
  - Download the .wyu file (specified in the .wys file)
  - Check the signature of the update
  - Apply the update

## Commands

- `cmd/updater` updater executable
- `cmd/wycparser` executable for parsing WYC files
- `cmd/wysparser` executable for parsing WYS files
- `cmd/wyuparser` executable for parsing WYU files (specifically the updtdetails.udt inside the archive)
