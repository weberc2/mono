package dedup

import (
	xslices "dedup/pkg/slices"
	"errors"
	"fmt"
	"hash/adler32"
	"io"
	"os"
	"slices"
)

func Dedup(notify Notifier, directory string) error {
	files := NewFileIter(directory)

	notify.ScanningDirectory(directory)
	inos := make(map[uint64]struct{})
	var uniqueInos []File
	for file, err, ok := files.Next(); ok; file, err, ok = files.Next() {
		if err != nil {
			return err
		}

		if _, exists := inos[file.Ino]; exists {
			continue
		}
		inos[file.Ino] = struct{}{}
		uniqueInos = append(uniqueInos, file)
	}

	notify.CollectedUniqueInoFiles(len(uniqueInos))
	slices.SortFunc(uniqueInos, func(l, r File) int {
		if l.Size < r.Size {
			return 1
		}
		if l.Size > r.Size {
			return -1
		}
		return 0
	})

	sizeGroups := xslices.GroupBy(uniqueInos, func(l, r *File) bool {
		return l.Size == r.Size
	})

	nonUniqueSizes := slices.DeleteFunc(sizeGroups, func(group []File) bool {
		return len(group) < 2
	})

	notify.IgnoringUniqueSizes(len(sizeGroups) - len(nonUniqueSizes))

	for i, sizeGroup := range nonUniqueSizes {
		notify.ProcessingSizeGroup(nonUniqueSizes, i)
		if err := ProcessSizeGroup(notify, sizeGroup); err != nil {
			return err
		}
	}

	return nil
}

const debug = false

func ProcessSizeGroup(notify Notifier, sizeGroup []File) error {
	for i := range sizeGroup {
		if err := sizeGroup[i].ChecksumBoundingBlocks(); err != nil {
			return err
		}
	}

	slices.SortFunc(sizeGroup, func(l, r File) int {
		if l.FirstBlockChecksum < r.FirstBlockChecksum {
			return -1
		}
		if l.FirstBlockChecksum > r.FirstBlockChecksum {
			return 1
		}
		if l.FinalBlockChecksum < r.FinalBlockChecksum {
			return -1
		}
		if l.FinalBlockChecksum > r.FinalBlockChecksum {
			return 1
		}
		return 0
	})

	byChecksums := xslices.GroupBy(sizeGroup, func(l, r *File) bool {
		return l.FirstBlockChecksum == r.FirstBlockChecksum &&
			l.FinalBlockChecksum == r.FinalBlockChecksum
	})

	if debug {
		for _, group := range byChecksums {
			if len(group) < 1 {
				continue
			}
			fmt.Printf(
				"  checksum group (len=%d,first=%d,final=%d)\n",
				len(group),
				group[0].FirstBlockChecksum,
				group[0].FinalBlockChecksum,
			)
			for i := range group {
				fmt.Printf("    %s\n", group[i].Path)
			}
		}
	}

	nonUnique := slices.DeleteFunc(
		byChecksums,
		func(files []File) bool { return len(files) < 2 },
	)
	notify.IgnoringUniqueChecksums(
		sizeGroup[0].Size,
		len(byChecksums)-len(nonUnique),
		len(nonUnique),
	)

	for _, files := range nonUnique {
		group := Group{
			Size:               files[0].Size,
			FirstBlockChecksum: files[0].FirstBlockChecksum,
			FinalBlockChecksum: files[0].FinalBlockChecksum,
			Paths:              make([]string, len(files)),
		}
		for i := range files {
			group.Paths[i] = files[i].Path
		}
		if err := DedupGroup(notify, &group); err != nil {
			return err
		}
	}

	return nil
}

func DedupGroup(notify Notifier, group *Group) error {
	notify.ProcessingGroup(group)
	notify.ChecksummingFile(group.Paths[0])
	if err := ensureUniquePath(
		group.Paths,
		func(p *string) string { return *p },
	); err != nil {
		return err
	}
	firstChecksum, err := ChecksumFile(group.Paths[0])
	if err != nil {
		return err
	}

	for _, path := range group.Paths[1:] {
		notify.ChecksummingFile(path)
		checksum, err := ChecksumFile(path)
		if err != nil {
			return err
		}

		if checksum == firstChecksum {
			notify.RemovingDuplicateFile(group.Size, path)
			if err := ToLink(path, group.Paths[0]); err != nil {
				return err
			}
		}
	}

	return nil
}

const dryRun = false

func ToLink(linkFile, linkedFile string) (err error) {
	if dryRun {
		return nil
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf(
				"replacing duplicate file `%s` with link to file `%s`: %w",
				linkFile,
				linkedFile,
				err,
			)
		}
	}()

	backup := linkFile + ".dedup-backup"
	if err := os.Rename(linkFile, backup); err != nil {
		return fmt.Errorf("creating backup link: %w", err)
	}

	if err := os.Link(linkedFile, linkFile); err != nil {
		return fmt.Errorf("creating new link: %w", err)
	}

	if err := os.Remove(backup); err != nil {
		return fmt.Errorf("removing backup link: %w", err)
	}
	return nil
}

func ChecksumFile(path string) (checksum uint32, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("checksumming file `%s`: %w", path, err)
		}
	}()

	var file *os.File
	if file, err = os.Open(path); err != nil {
		err = fmt.Errorf("opening file: %w", err)
		return
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	hash := adler32.New()
	if _, err = io.Copy(hash, file); err != nil {
		err = fmt.Errorf("hashing file contents: %w", err)
		return
	}

	checksum = hash.Sum32()
	return
}

func ensureUniquePath[T any](group []T, pathfn func(*T) string) error {
	if debug {
		seen := make(map[string]struct{})
		for i := range group {
			path := pathfn(&group[i])
			if _, exists := seen[path]; exists {
				return fmt.Errorf("duplicate path: %s", path)
			}
			seen[path] = struct{}{}
		}
	}
	return nil
}
