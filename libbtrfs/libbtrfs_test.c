#include "btrfs.h"
#include <errno.h>
#include <string.h>
#include <stdio.h>
int main() {
	int ret = 0;
	const char* subvol_path = "/media/jarekdam/e52c00b9-60b2-468a-83cc-e6c652f098f7/subvol";
	const char* snapshot_path = "/media/jarekdam/e52c00b9-60b2-468a-83cc-e6c652f098f7/subvol_snap";
	if ((ret = btrfs_create_subvol(subvol_path))) {
		printf("Error when creating subvol: %s\n", strerror(-ret));
	}

	if ((ret = btrfs_create_snapshot(subvol_path, snapshot_path))) {
		printf("Error when creating snapshot: %s\n", strerror(-ret));
	}

	if ((ret = btrfs_delete_subvol(subvol_path))) {
		printf("Error when deleting subvol: %s\n", strerror(-ret));
	}
	return 0;
}