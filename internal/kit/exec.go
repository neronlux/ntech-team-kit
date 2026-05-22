package kit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RunInstaller executes the install.sh script with the given arguments.
// It sets the working directory to the kit root so relative paths in the script work.
func RunInstaller(kitRoot string, args []string) error {
	script := filepath.Join(kitRoot, "install.sh")

	cmd := exec.Command(script, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = kitRoot

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install.sh failed: %w", err)
	}
	return nil
}
