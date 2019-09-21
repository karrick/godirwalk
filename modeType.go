package godirwalk

import (
	"os"
)

func modeTypeLStat(osPathname string) (os.FileMode, error) {
	fi, err := os.Lstat(osPathname)
	if err == nil {
		// Even though the stat provided all file mode bits, we want to ensure
		// same values returned to caller regardless of whether we obtained file
		// mode bits from syscall or stat call.  Therefore mask out the
		// additional file mode bits that are provided by stat but not by the
		// syscall, so users can rely on their values.
		return fi.Mode() & os.ModeType, nil
	}
	return 0, err
}
