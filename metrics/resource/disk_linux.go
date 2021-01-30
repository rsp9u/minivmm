package resource

import "golang.org/x/sys/unix"

func getDiskSize(path string) uint64 {
	var stat unix.Statfs_t
	unix.Statfs(path, &stat)

	return stat.Blocks * uint64(stat.Bsize)
}
