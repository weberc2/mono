package dedup

import (
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
)

type Notifier struct {
	w io.Writer
}

func NewNotifier(w io.Writer) (n Notifier) {
	n.w = w
	return
}

func (n Notifier) ScanningDirectory(directory string) {
	fmt.Fprintf(
		n.w,
		"%s scanning directory: %s\n",
		nowStr(),
		directory,
	)
}

func (n Notifier) CollectedUniqueInoFiles(count int) {
	fmt.Fprintf(
		n.w,
		"%s collected %d files with distinct inos\n",
		nowStr(),
		count,
	)
}

func (n Notifier) IgnoringUniqueSizes(ignored int) {
	fmt.Fprintf(
		n.w,
		"%s ignoring %d files with unique sizes\n",
		nowStr(),
		ignored,
	)
}

var bold = color.New(color.Bold)

func (n Notifier) ProcessingSizeGroup(groups [][]File, index int) {
	bold.Fprintf(
		n.w,
		"\n%s processing size group %d/%d (%d files @ %s each)\n",
		nowStr(),
		index+1,
		len(groups),
		len(groups[index]),
		human(groups[index][0].Size),
	)
}

func (n Notifier) IgnoringUniqueChecksums(size int64, ignored, remaining int) {
	if ignored < 1 {
		return
	}
	color.Green(
		"%s  ignoring %d files with unique checksums (%d groups remaining)\n",
		nowStr(),
		ignored,
		remaining,
	)
}

func (n Notifier) ProcessingGroup(group *Group) {
	bold.Fprintf(
		n.w,
		"%s  processing group (%d files @ %s each)\n",
		nowStr(),
		len(group.Paths),
		human(group.Size),
	)
}

func (n Notifier) ChecksummingFile(path string) {
	fmt.Fprintf(n.w, "%s    checksumming file [%s]\n", nowStr(), path)
}

func (n Notifier) RemovingDuplicateFile(size int64, path string) {
	color.Green(
		"%s    removing duplicate file (size: %s) [%s]\n",
		nowStr(),
		human(size),
		path,
	)
}

func human(n int64) string {
	// Metric suffixes
	const (
		K = 1_000
		M = 1_000_000
		G = 1_000_000_000
		T = 1_000_000_000_000
		P = 1_000_000_000_000_000
		E = 1_000_000_000_000_000_000
	)

	// Handle negative numbers
	neg := ""
	if n < 0 {
		neg = "-"
		n = -n
	}

	switch {
	case n >= E:
		return fmt.Sprintf("%s%.0fE", neg, float64(n)/E)
	case n >= P:
		return fmt.Sprintf("%s%.0fP", neg, float64(n)/P)
	case n >= T:
		return fmt.Sprintf("%s%.0fT", neg, float64(n)/T)
	case n >= G:
		return fmt.Sprintf("%s%.0fG", neg, float64(n)/G)
	case n >= M:
		return fmt.Sprintf("%s%.0fM", neg, float64(n)/M)
	case n >= K:
		return fmt.Sprintf("%s%.0fK", neg, float64(n)/K)
	default:
		return fmt.Sprintf("%s%d", neg, n)
	}
}

func nowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
