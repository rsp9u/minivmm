package resource

import "golang.org/x/sys/unix"

func getDiskSizeTotal(path string) uint64 {
	var stat unix.Statfs_t
	unix.Statfs(path, &stat)

	return stat.Blocks * uint64(stat.Bsize)
}

func getDiskSizeUsed(path string) uint64 {
	var stat unix.Statfs_t
	unix.Statfs(path, &stat)

	return (stat.Blocks - stat.Bavail) * uint64(stat.Bsize)
}
