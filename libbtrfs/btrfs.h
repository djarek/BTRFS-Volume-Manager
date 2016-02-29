#ifndef BTRFS_H
#define BTRFS_H

int btrfs_create_snapshot(const char *source_path, const char *dst_path);
int btrfs_delete_subvol(const char *path);
int btrfs_create_subvol(const char *path);

#endif // BTRFS_H