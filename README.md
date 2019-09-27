# win-service-updater

![](https://github.com/huntresslabs/win-service-updater/workflows/Build/badge.svg)
![](https://github.com/huntresslabs/win-service-updater/workflows/Test/badge.svg)

Partial implementation of wyUpdate functionality. This updater is written in GoLang to avoid .NET dependencies.

## Goals

- Compatibility with existing wyUpdate binary files
- Drop in replacement for use in existing service update commands (should work for any update, there is just no GUI)

## Differences

- designed to only be run from a service or command-line, there is no GUI component
- only full binary replacement (no diff)
- only supports stopping/starting services before/after update
  - no registry updates, COM updates, etc.
- no functionality to elevate privileges (need admin to install a service anyway)

## Arguments supported

- "/quickcheck"
- "/justcheck"
- "/noerr",
- "-urlargs=_args_"
- "/outputinfo=_out_"
- "/fromservice"
- "-logfile=_log_"
- "-cdata=_file_"
- "-server=_url_"

## Operation

- To check if an update is available:
  - Download the .wys file (update URL specified in client.wyc)
  - Compare the available update (speficied in the .wys file) with the version currently installed
- If an update is required:
  - Download the .wyu file (specified in the .wys file)
  - Check the signature of the update
  - Apply the update
