package localrunner

import (
	"fmt"
)

var Verbose = false

func local_log_info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf(" [\x1b[33m~~\x1b[m] %s\n", msg)
}

func local_log_verbose(format string, args ...any) {
	if Verbose {
		msg := fmt.Sprintf(format, args...)
		fmt.Printf(" [\x1b[34mvv\x1b[m] %s\n", msg)
	}
}

func local_log_error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf(" [\x1b[31m!!\x1b[m] %s\n", msg)
}

func local_log_success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf(" [\x1b[32m**\x1b[m] %s\n", msg)
}
