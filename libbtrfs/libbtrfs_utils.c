#include "ioctl.h"
#include <stdlib.h>
#include <linux/limits.h>
#include <linux/magic.h>
#include <unistd.h>
#include <dirent.h>
#include <stdlib.h>
#include "kerncompat.h"

#define BTRFS_SUPER_INFO_SIZE 4096
#define BTRFS_SUPER_INFO_OFFSET (64 * 1024)

/*
 * For a given path, fill in the ioctl fs_ and info_ args.
 * If the path is a btrfs mountpoint, fill info for all devices.
 * If the path is a btrfs device, fill in only that device.
 *
 * The path provided must be either on a mounted btrfs fs,
 * or be a mounted btrfs device.
 *
 * Returns 0 on success, or a negative errno.
 */
int get_fs_info(char *path, struct btrfs_ioctl_fs_info_args *fi_args,
		struct btrfs_ioctl_dev_info_args **di_ret)
{
	int fd = -1;
	int ret = 0;
	int ndevs = 0;
	int i = 0;
	int replacing = 0;
	struct btrfs_fs_devices *fs_devices_mnt = NULL;
	struct btrfs_ioctl_dev_info_args *di_args;
	struct btrfs_ioctl_dev_info_args tmp;
	char mp[PATH_MAX];
	DIR *dirstream = NULL;

	memset(fi_args, 0, sizeof(*fi_args));

	if (is_block_device(path) == 1) {
		struct btrfs_super_block *disk_super;
		char buf[BTRFS_SUPER_INFO_SIZE];
		u64 devid;

		/* Ensure it's mounted, then set path to the mountpoint */
		fd = open(path, O_RDONLY);
		if (fd < 0) {
			ret = -errno;
			fprintf(stderr, "Couldn't open %s: %s\n",
				path, strerror(errno));
			goto out;
		}
		ret = check_mounted_where(fd, path, mp, sizeof(mp),
					  &fs_devices_mnt);
		if (!ret) {
			ret = -EINVAL;
			goto out;
		}
		if (ret < 0)
			goto out;
		path = mp;
		/* Only fill in this one device */
		fi_args->num_devices = 1;

		disk_super = (struct btrfs_super_block *)buf;
		ret = btrfs_read_dev_super(fd, disk_super,
					   BTRFS_SUPER_INFO_OFFSET, 0);
		if (ret < 0) {
			ret = -EIO;
			goto out;
		}
		devid = btrfs_stack_device_id(&disk_super->dev_item);

		fi_args->max_id = devid;
		i = devid;

		memcpy(fi_args->fsid, fs_devices_mnt->fsid, BTRFS_FSID_SIZE);
		close(fd);
	}

	/* at this point path must not be for a block device */
	fd = open_file_or_dir(path, &dirstream);
	if (fd < 0) {
		ret = -errno;
		goto out;
	}

	/* fill in fi_args if not just a single device */
	if (fi_args->num_devices != 1) {
		ret = ioctl(fd, BTRFS_IOC_FS_INFO, fi_args);
		if (ret < 0) {
			ret = -errno;
			goto out;
		}

		/*
		 * The fs_args->num_devices does not include seed devices
		 */
		ret = search_chunk_tree_for_fs_info(fd, fi_args);
		if (ret)
			goto out;

		/*
		 * search_chunk_tree_for_fs_info() will lacks the devid 0
		 * so manual probe for it here.
		 */
		ret = get_device_info(fd, 0, &tmp);
		if (!ret) {
			fi_args->num_devices++;
			ndevs++;
			replacing = 1;
			if (i == 0)
				i++;
		}
	}

	if (!fi_args->num_devices)
		goto out;

	di_args = *di_ret = malloc((fi_args->num_devices) * sizeof(*di_args));
	if (!di_args) {
		ret = -errno;
		goto out;
	}

	if (replacing)
		memcpy(di_args, &tmp, sizeof(tmp));
	for (; i <= fi_args->max_id; ++i) {
		ret = get_device_info(fd, i, &di_args[ndevs]);
		if (ret == -ENODEV)
			continue;
		if (ret)
			goto out;
		ndevs++;
	}

	/*
	* only when the only dev we wanted to find is not there then
	* let any error be returned
	*/
	if (fi_args->num_devices != 1) {
		BUG_ON(ndevs == 0);
		ret = 0;
	}

out:
	close_file_or_dir(fd, dirstream);
	return ret;
}