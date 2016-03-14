
#include "btrfs-iface/ioctl.h"
#include "btrfs.h"
#include <fcntl.h>
#include <unistd.h>
#include <string.h>
#include <errno.h>
#include <dirent.h>
#include <sys/statfs.h>
#include <sys/stat.h>
#include <sys/ioctl.h>
#include <linux/magic.h>
#include <stdio.h>
#include <stdlib.h>
#include <libgen.h>

#include <blkid/blkid.h>

int btrfs_create_subvol(const char *path)
{
	int ret = 0;
	char *name = NULL, *root_path = NULL;

	struct btrfs_ioctl_vol_args vol_args;
	memset(&vol_args, 0, sizeof(vol_args));

	char* path_dname = strdup(path);
	root_path = realpath(dirname(path_dname), NULL);
	if (!root_path) {
		ret = -errno;
		goto out;
	}
	name = strdup(path);

	const char *subvol_name = basename(name);

	int fd = open(root_path, O_DIRECTORY);
	if (fd == -1) {
		ret = -errno;
		goto out;
	}
	strncpy(vol_args.name, subvol_name, BTRFS_SUBVOL_NAME_MAX);

	if (ioctl(fd, BTRFS_IOC_SUBVOL_CREATE, &vol_args)) {
		ret = -errno;
		goto out;
	}

out:
	close(fd);
	free(name);
	free(root_path);
	free(path_dname);
	return ret;
}

int btrfs_delete_subvol(const char *path)
{
	int ret = 0;
	int fd = -1;
	char *name = NULL;

	struct btrfs_ioctl_vol_args vol_args;
	memset(&vol_args, 0, sizeof(vol_args));

	char* path_dname = strdup(path);
	char* root_path = realpath(dirname(path_dname), NULL);
	if (!root_path) {
		ret = -errno;
		goto out;
	}
	name = strdup(path);
	const char *subvol_name = basename(name);

	fd = open(root_path, O_DIRECTORY);
	if (fd == -1) {
		ret = -errno;
		goto out;
	}

	strncpy(vol_args.name, subvol_name, BTRFS_SUBVOL_NAME_MAX);

	if (ioctl(fd, BTRFS_IOC_SNAP_DESTROY, &vol_args)) {
		ret = -errno;
		goto out;
	}

out:
	close(fd);
	free(path_dname);
	free(name);
	free(root_path);
	return ret;
}

int btrfs_create_snapshot(const char *source_path, const char *dst_path)
{
	int ret = 0;
	char *name = NULL;

	struct btrfs_ioctl_vol_args_v2 vol_args;
	memset(&vol_args, 0, sizeof(vol_args));

	char *path_dname = strdup(dst_path);
	char *dst_root_path = realpath(dirname(path_dname), NULL);
	if (!path_dname) {
		ret = -errno;
		goto out;
	}

	name = strdup(dst_path);
	const char *subvol_name = basename(name);

	int dst_fd = -1;
	int source_fd = open(source_path, O_DIRECTORY);

	if (source_fd == -1) {
		ret = -errno;
		goto out;
	}

	dst_fd = open(dst_root_path, O_DIRECTORY);
	if (dst_fd == -1) {
		ret = -errno;
		goto out;
	}

	vol_args.fd = source_fd;
	strncpy(vol_args.name, subvol_name, BTRFS_SUBVOL_NAME_MAX);

	if (ioctl(dst_fd, BTRFS_IOC_SNAP_CREATE_V2, &vol_args)) {
		ret = -errno;
		goto out;
	}

out:
	close(source_fd);
	close(dst_fd);
	free(path_dname);
	free(dst_root_path);
	free(name);
	return ret;
}

void blkid_dev_to_block_device(blkid_dev dev, struct block_device *bd) {
	bd->dev_name = strdup(blkid_dev_devname(dev));
	blkid_tag_iterate tag_iterator = blkid_tag_iterate_begin(dev);
	const char *type, *value;
	while (blkid_tag_next(tag_iterator, &type, &value) == 0) {
		if (strcmp(type, "UUID") == 0) {
			bd->UUID = strdup(value);	
		}
		if (strcmp(type, "TYPE") == 0) {
			bd->type = strdup(value);
		}
	}
	
	blkid_tag_iterate_end(tag_iterator);
}

int get_devices(struct block_devices_array *arr)
{
	blkid_cache cache = NULL;
	blkid_dev_iterate iter = NULL;
	blkid_dev dev = NULL;
	int ret = 0;
	arr->count = 0;
	arr->devs = NULL;

	if (arr == NULL) {
		return -1;
	}

	if (blkid_get_cache(&cache, NULL) < 0) {
		return -1;
	}
	blkid_probe_all(cache);
	iter = blkid_dev_iterate_begin(cache);

	while (blkid_dev_next(iter, &dev) == 0) {
		dev = blkid_verify(cache, dev);
		if (dev) {
			arr->devs = realloc(arr->devs, (++arr->count)*sizeof(struct block_device));
			struct block_device *bd =  &arr->devs[arr->count - 1];
			memset(bd, 0, sizeof(struct block_device));

			blkid_dev_to_block_device(dev, bd);
		}
	}

	blkid_dev_iterate_end(iter);
	blkid_put_cache(cache);
	return ret;

}

void block_device_free(struct block_device *dev) {
	free(dev->dev_name);
	free(dev->type);
	free(dev->UUID);
}

void block_devices_array_free(struct block_devices_array arr) {
	for (int i = 0; i < arr.count; ++i) {
		block_device_free(&arr.devs[i]);
	}
	free(arr.devs);
}
