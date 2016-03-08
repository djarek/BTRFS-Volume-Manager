#ifndef BTRFS_H
#define BTRFS_H

int btrfs_create_snapshot(const char *source_path, const char *dst_path);
int btrfs_delete_subvol(const char *path);
int btrfs_create_subvol(const char *path);

struct block_device
{
	char *dev_name;
	char *UUID;
	char *type;
};

struct block_devices_array
{
	struct block_device *devs;
	int count;

};

void block_device_free(struct block_device *dev);
void block_devices_array_free(struct block_devices_array arr);

int get_devices(struct block_devices_array*);

#endif // BTRFS_H