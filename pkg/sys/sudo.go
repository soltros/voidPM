package sys

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// IsRoot checks if current user has effective UID 0
func IsRoot() bool {
	return os.Geteuid() == 0
}

// EnsurePrivileges checks if current user is root, returning error if unprivileged
func EnsurePrivileges() error {
	if !IsRoot() {
		return fmt.Errorf("this operation requires root privileges; rerun with 'sudo' or elevated permissions")
	}
	return nil
}

// FindElevator returns "doas" or "sudo" depending on availability, or empty if none found
func FindElevator() string {
	if _, err := exec.LookPath("doas"); err == nil {
		return "doas"
	}
	if _, err := exec.LookPath("sudo"); err == nil {
		return "sudo"
	}
	return ""
}

// WrapElevated wraps a command slice with sudo or doas if not running as root
func WrapElevated(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no command provided for execution")
	}

	if IsRoot() {
		return args, nil
	}

	elevator := FindElevator()
	if elevator == "" {
		return nil, fmt.Errorf("root privileges required for '%s'. Please run as root or install sudo/doas", strings.Join(args, " "))
	}

	return append([]string{elevator}, args...), nil
}

// RunElevated executes a command with elevated privileges, connecting stdin/stdout/stderr for interactive prompt
func RunElevated(args ...string) error {
	cmdArgs, err := WrapElevated(args)
	if err != nil {
		return err
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// RunElevatedCombined runs command and returns combined output
func RunElevatedCombined(args ...string) (string, error) {
	cmdArgs, err := WrapElevated(args)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
