package deps

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// PlatformInstaller handles platform-specific installations
type PlatformInstaller struct {
	OS             string
	Architecture   string
	PackageManager string
}

// NewPlatformInstaller creates a new platform installer
func NewPlatformInstaller() *PlatformInstaller {
	return &PlatformInstaller{
		OS:             runtime.GOOS,
		Architecture:   runtime.GOARCH,
		PackageManager: DetectPackageManager(),
	}
}

// InstallWheresmyprompt installs wheresmyprompt using the best available method
func (p *PlatformInstaller) InstallWheresmyprompt() (*InstallationResult, error) {
	switch p.PackageManager {
	case "homebrew":
		return p.installViaHomebrew("toozej/tap/wheresmyprompt")
	case "go":
		return p.installViaGo("github.com/toozej/wheresmyprompt/cmd/wheresmyprompt@latest")
	default:
		return p.installBinaryFromGitHub("toozej", "wheresmyprompt", "wheresmyprompt")
	}
}

// InstallFiles2prompt installs files2prompt using the best available method
func (p *PlatformInstaller) InstallFiles2prompt() (*InstallationResult, error) {
	switch p.PackageManager {
	case "homebrew":
		return p.installViaHomebrew("toozej/tap/files2prompt")
	case "go":
		return p.installViaGo("github.com/toozej/files2prompt/cmd/files2prompt@latest")
	default:
		return p.installBinaryFromGitHub("toozej", "files2prompt", "files2prompt")
	}
}

// InstallLLM installs llm CLI using the best available method
func (p *PlatformInstaller) InstallLLM() (*InstallationResult, error) {
	switch p.PackageManager {
	case "homebrew":
		return p.installViaHomebrew("llm")
	case "pipx":
		return p.installViaPipx("llm")
	case "pip":
		return p.installViaPip("llm")
	default:
		// Try to install pipx first, then llm
		if err := p.ensurePipxInstalled(); err != nil {
			return &InstallationResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to install pipx: %v", err),
			}, err
		}
		return p.installViaPipx("llm")
	}
}

// installViaHomebrew installs a package using Homebrew
func (p *PlatformInstaller) installViaHomebrew(packageName string) (*InstallationResult, error) {
	cmd := exec.Command("brew", "install", packageName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &InstallationResult{
			Success: false,
			Error:   fmt.Sprintf("Homebrew installation failed: %v\nOutput: %s", err, string(output)),
		}, err
	}

	return &InstallationResult{
		Success: true,
		Message: fmt.Sprintf("Successfully installed %s via Homebrew", packageName),
	}, nil
}

// installViaGo installs a package using go install
func (p *PlatformInstaller) installViaGo(packagePath string) (*InstallationResult, error) {
	// Check if Go is available
	if _, err := exec.LookPath("go"); err != nil {
		return &InstallationResult{
			Success: false,
			Error:   "Go toolchain not found - please install Go first",
		}, err
	}

	cmd := exec.Command("go", "install", packagePath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &InstallationResult{
			Success: false,
			Error:   fmt.Sprintf("Go install failed: %v\nOutput: %s", err, string(output)),
		}, err
	}

	return &InstallationResult{
		Success: true,
		Message: fmt.Sprintf("Successfully installed %s via Go", packagePath),
	}, nil
}

// installViaPipx installs a package using pipx
func (p *PlatformInstaller) installViaPipx(packageName string) (*InstallationResult, error) {
	cmd := exec.Command("pipx", "install", packageName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &InstallationResult{
			Success: false,
			Error:   fmt.Sprintf("pipx installation failed: %v\nOutput: %s", err, string(output)),
		}, err
	}

	return &InstallationResult{
		Success: true,
		Message: fmt.Sprintf("Successfully installed %s via pipx", packageName),
	}, nil
}

// installViaPip installs a package using pip
func (p *PlatformInstaller) installViaPip(packageName string) (*InstallationResult, error) {
	cmd := exec.Command("pip", "install", packageName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &InstallationResult{
			Success: false,
			Error:   fmt.Sprintf("pip installation failed: %v\nOutput: %s", err, string(output)),
		}, err
	}

	return &InstallationResult{
		Success: true,
		Message: fmt.Sprintf("Successfully installed %s via pip", packageName),
	}, nil
}

// installBinaryFromGitHub downloads and installs a binary from GitHub releases
func (p *PlatformInstaller) installBinaryFromGitHub(owner, repo, binaryName string) (*InstallationResult, error) {
	// Get the latest release URL
	downloadURL, err := p.getGitHubReleaseURL(owner, repo)
	if err != nil {
		return &InstallationResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to get release URL: %v", err),
		}, err
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s-*.tar.gz", binaryName))
	if err != nil {
		return &InstallationResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create temp file: %v", err),
		}, err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Download the binary
	fmt.Printf("Downloading %s from %s...\n", binaryName, downloadURL)
	if err := p.downloadFile(downloadURL, tempFile); err != nil {
		return &InstallationResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to download binary: %v", err),
		}, err
	}

	// Extract and install
	installPath := p.getInstallPath(binaryName)
	if err := p.extractAndInstall(tempFile.Name(), binaryName, installPath); err != nil {
		return &InstallationResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to extract and install: %v", err),
		}, err
	}

	return &InstallationResult{
		Success: true,
		Message: fmt.Sprintf("Successfully installed %s to %s", binaryName, installPath),
	}, nil
}

// getGitHubReleaseURL constructs the GitHub release download URL
func (p *PlatformInstaller) getGitHubReleaseURL(owner, repo string) (string, error) {
	osName := p.OS
	if osName == "darwin" {
		osName = "macOS"
	}

	arch := p.Architecture
	if arch == "amd64" {
		arch = "x86_64"
	} else if arch == "arm64" {
		arch = "arm64"
	}

	// GitHub releases typically follow this pattern
	filename := fmt.Sprintf("%s_%s_%s.tar.gz", repo, osName, arch)
	url := fmt.Sprintf("https://github.com/%s/%s/releases/latest/download/%s", owner, repo, filename)

	return url, nil
}

// downloadFile downloads a file from URL to the given writer with progress indication
func (p *PlatformInstaller) downloadFile(url string, dest io.Writer) error {
	resp, err := http.Get(url) // #nosec G107 -- URL from trusted dependency configuration
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Simple progress indication
	fmt.Printf("Download in progress...")
	_, err = io.Copy(dest, resp.Body)
	if err == nil {
		fmt.Println(" completed!")
	}

	return err
}

// extractAndInstall extracts a tar.gz file and installs the binary
func (p *PlatformInstaller) extractAndInstall(archivePath, binaryName, installPath string) error {
	// Open the archive
	file, err := os.Open(archivePath) // #nosec G304 -- Archive path is from validated dependency configuration
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Extract the binary
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Look for the binary file
		if header.Typeflag == tar.TypeReg && strings.Contains(header.Name, binaryName) {
			// Check file size to prevent decompression bomb
			const maxFileSize = 100 * 1024 * 1024 // 100MB limit
			if header.Size > maxFileSize {
				return fmt.Errorf("file %s is too large (%d bytes > %d bytes limit)", header.Name, header.Size, maxFileSize)
			}

			// Ensure install directory exists
			installDir := filepath.Dir(installPath)
			if err := os.MkdirAll(installDir, 0750); err != nil {
				return fmt.Errorf("failed to create install directory: %w", err)
			}

			// Create the binary file
			outFile, err := os.OpenFile(installPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0750) // #nosec G304 G302 -- Install path from validated config, 0750 appropriate for executables
			if err != nil {
				return fmt.Errorf("failed to create binary file: %w", err)
			}

			// Copy binary content with size limit
			_, err = io.CopyN(outFile, tarReader, header.Size)
			if closeErr := outFile.Close(); closeErr != nil {
				return fmt.Errorf("failed to close binary file: %w", closeErr)
			}
			if err != nil {
				return fmt.Errorf("failed to write binary: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("binary %s not found in archive", binaryName)
}

// getInstallPath returns the appropriate install path for a binary
func (p *PlatformInstaller) getInstallPath(binaryName string) string {
	// Try to use ~/bin if it exists, otherwise use /usr/local/bin on Unix systems
	homeDir, err := os.UserHomeDir()
	if err == nil {
		binDir := filepath.Join(homeDir, "bin")
		if _, err := os.Stat(binDir); err == nil {
			return filepath.Join(binDir, binaryName)
		}

		// Create ~/bin if it doesn't exist
		if err := os.MkdirAll(binDir, 0750); err == nil {
			return filepath.Join(binDir, binaryName)
		}
	}

	// Fallback to system paths
	if p.OS == "windows" {
		return filepath.Join("C:", "tools", "bin", binaryName+".exe")
	}

	return filepath.Join("/usr/local/bin", binaryName)
}

// ensurePipxInstalled ensures pipx is installed on the system
func (p *PlatformInstaller) ensurePipxInstalled() error {
	// Check if pipx is already available
	if _, err := exec.LookPath("pipx"); err == nil {
		return nil
	}

	// Try to install pipx using the platform package manager
	switch p.PackageManager {
	case "homebrew":
		cmd := exec.Command("brew", "install", "pipx")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install pipx via homebrew: %w", err)
		}
	case "apt":
		cmd := exec.Command("sudo", "apt", "install", "-y", "pipx")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install pipx via apt: %w", err)
		}
	case "pip":
		cmd := exec.Command("pip", "install", "--user", "pipx")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install pipx via pip: %w", err)
		}
	default:
		return fmt.Errorf("cannot install pipx: no supported package manager found")
	}

	return nil
}

// VerifyInstallation verifies that a binary was installed correctly
func (p *PlatformInstaller) VerifyInstallation(binaryName string) error {
	_, err := exec.LookPath(binaryName)
	if err != nil {
		return fmt.Errorf("binary %s not found in PATH after installation", binaryName)
	}
	return nil
}

// UpdatePathEnvironment adds a directory to PATH if not already present
func (p *PlatformInstaller) UpdatePathEnvironment(newPath string) error {
	// This is a simplified implementation
	// In practice, you'd want to update shell rc files
	currentPath := os.Getenv("PATH")

	separator := ":"
	if p.OS == "windows" {
		separator = ";"
	}

	if !strings.Contains(currentPath, newPath) {
		newPathEnv := newPath + separator + currentPath
		return os.Setenv("PATH", newPathEnv)
	}

	return nil
}
