# win-service-updater

![](https://github.com/huntresslabs/win-service-updater/workflows/Build/badge.svg)
![](https://github.com/huntresslabs/win-service-updater/workflows/Test/badge.svg)

Implementation of "core" wyUpdate functionality. wyUpdate is a utility written in .NET for updating Windows applications. We used wyUpdate dependably for years, but couldn't rely on a standard version of .NET across our customer install base. Ultimately we decided to reimplement the wyUpdate functionality we used in GoLang.

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

## Commands

- Build `cmd/updater` for main updater executable
- Build `cmd/wycparser` for WYC parser executable (specifically the iuclient.iuc inside the archive)
- Build `cmd/wysparser` for WYS parser executable
- Build `cmd/wyuparser` for WYU parser executable (specifically the updtdetails.udt inside the archive)

## General Operation

- To check if an update is available:
  - Download the .wys file (URL specified in client.wyc)
  - Compare the available update (specified in the .wys file) with the version currently installed
- If an update is required:
  - Download the .wyu file (URL specified in the .wys file)
  - If the update is signed, verify the signature of the update otherwise verify the checksum
  - Apply the update
  - Update the version number contained within the client.wyc
