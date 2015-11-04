# Docker volume LVM snapshot mounter

Docker volume extension that creates a snapshot of an existing Logical Volume, mounts it,
exposes it to the container. Discards snapshotted volumes after container is stopped.

Useful for eg. providing pre-seeded MySQL databases to containers for testing purposes.

Please note that XFS filesystems do not like to be mounted simultaneously on the same
system: http://www.miljan.org/main/2009/11/16/lvm-snapshots-and-xfs/

## Requirements

Docker 1.8.3 with the patch from: https://patch-diff.githubusercontent.com/raw/docker/docker/pull/14737.patch

## Usage

`make build` to build the container which contains the plugin binary.

`make containerrun` to run the volume plugin in a container, listening to the socket in the default
`/run/docker/plugins/` dir.

To use the plugin to create a snapshot of /dev/volume_group/logical_volume and mount
it to a container, run:

`docker run --rm -it --volume-driver=snapshot -v volume_group/logical_volume:/mnt busybox ls -la`

## Build and run in Boot2Docker qemu VM

`make iso` will build an experimental boot2docker ISO which will auto-start the snapshot volume-plugin.

If you're on Linux, you can run it in Qemu-kvm using `make run`.

## License

MIT
