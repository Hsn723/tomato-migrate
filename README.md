# tomato-migrate

## Background

There are two methods to save/restore configurations on a router with the Tomato firmware installed. These methods however are meant to be used on the same router, but there are situations where one might want to restore to a different router. For instance, the original router might have a hardware failure and no longer powering on.

The first method involves dumping the `nvram` and restoring specific keys. This method is however not user-friendly and requires an understanding of what each configuration variable does, which is something that even most power users may not fully grasp.

The other method involves using the backup/restore menu in the GUI, which generates a `.cfg` file. A Tomato `.cfg` file is a gzipped (compressed) text file which store settings using the null byte (`\0`) as the end-of-line symbol. For this reason, it is not always straightforward to open and edit, especially on Windows systems.

With both methods, it is not possible to restore using the entire file as-is, as some of the settings, namely hardware MAC address mappings, are hardware-dependent. We can read through the `nvram` backup and restore the settings one by one, ignoring those pertaining to MAC addresses, but that manual process is tedious and error-prone. We can manually edit the Tomato `.cfg` file by uncompressing it and opening it in an appropriate text editor, but this is not user-friendly and error-prone.

`tomato-migrate` targets the more user-friendly `.cfg` restore method, and enables restoring to another Tomato of the same model and firmware.

## Installation

On systems with a recent version of `go` installed, `go install` can be used.

```sh
go install github.com/hsn723/tomato-migrate@latest
```

Pre-compiled binaries are also available in the Releases page.


## Usage

```
tomato-migrate -i FILE -o FILE
```

The following assumes you are in a disaster recovery situation where the original router no longer works and you are aiming to restore an existing `.cfg` backup (referred to as `old.cfg` going forward) to a new Tomato router of the exact same model and firmware.

- Perform a fresh install of Tomato on the new router
- Download a backup of the new router's configuration (referred to as `new.cfg` going forward)
- Run `tomato-migrate` providing it with the paths to `old.cfg` and `new.cfg`, with `old.cfg` as the input and `new.cfg` as the output
    - Example: `tomato-migrate -i old.cfg -o new.cfg
- `tomato-migrate` will
    - Create a backup of `new.cfg`
    - Read `new.cfg` and get a list of the new hardware MAC addresses
    - Read `old.cfg` and remap MAC addresses to those of `new.cfg`
    - Write the resulting configuration to `new.cfg`
- You can now upload the resulting `new.cfg` to the new Tomato router to restore the old router's configuration
