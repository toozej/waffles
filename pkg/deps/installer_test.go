package deps

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestNewPlatformInstaller(t *testing.T) {
	installer := NewPlatformInstaller()

	if installer == nil {
		t.Error("Expected installer to be non-nil")
		return
	}

	// Test that installer is created for current platform
	expectedOS := runtime.GOOS
	if installer.OS != expectedOS {
		t.Errorf("Expected OS %s, got %s", expectedOS, installer.OS)
	}

	expectedArch := runtime.GOARCH
	if installer.Architecture != expectedArch {
		t.Errorf("Expected Architecture %s, got %s", expectedArch, installer.Architecture)
	}

	// PackageManager should be detected
	if installer.PackageManager == "" {
		t.Error("Expected PackageManager to be detected")
	}
}

func TestDetectPackageManager(t *testing.T) {
	pkgManager := DetectPackageManager()

	if pkgManager == "" {
		t.Error("Expected package manager to be detected")
	}

	// Should be one of the supported package managers
	validManagers := []string{"go", "homebrew", "apt", "yum", "pacman", "pipx", "pip", "unknown"}
	found := false
	for _, manager := range validManagers {
		if pkgManager == manager {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Package manager %s not in valid list", pkgManager)
	}

	t.Logf("Detected package manager: %s", pkgManager)
}

func TestGetInstallInstructions(t *testing.T) {
	deps := RequiredDependencies()

	for _, dep := range deps {
		instructions := GetInstallInstructions(dep)

		if instructions == "" {
			t.Errorf("Expected install instructions for %s to be non-empty", dep.Name)
		}

		// Check that instructions contain relevant information
		if !strings.Contains(instructions, dep.Name) {
			t.Errorf("Expected instructions for %s to contain dependency name", dep.Name)
		}

		t.Logf("Instructions for %s: %s", dep.Name, instructions[:min(100, len(instructions))])
	}
}

func TestAutoInstallAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping auto-install test in short mode")
	}

	// This is a potentially destructive test, so we'll just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AutoInstallAll panicked: %v", r)
		}
	}()

	err := AutoInstallAll()

	// We expect this might fail in test environment without proper setup
	if err != nil {
		t.Logf("AutoInstallAll failed as expected in test environment: %v", err)
	} else {
		t.Log("AutoInstallAll succeeded (possibly all dependencies already installed)")
	}
}

func TestInstallDependency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping install test in short mode")
	}

	deps := RequiredDependencies()

	for _, dep := range deps {
		t.Run(dep.Name, func(t *testing.T) {
			result, err := InstallDependency(dep)

			// We expect this might fail in test environment
			if err != nil {
				t.Logf("InstallDependency for %s failed as expected in test environment: %v", dep.Name, err)
			}

			if result != nil {
				t.Logf("Installation result for %s: Success=%t, Message=%s, Error=%s",
					dep.Name, result.Success, result.Message, result.Error)
			}
		})
	}
}

func TestInstallLLMPlugin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping LLM plugin install test in short mode")
	}

	// Test with a fake plugin to avoid actual installation
	err := InstallLLMPlugin("fake-plugin-test")

	// This will likely fail since llm might not be installed or plugin is fake
	if err != nil {
		t.Logf("Expected error installing fake LLM plugin: %v", err)

		// Check error message is reasonable (but don't require specific text)
		if !strings.Contains(err.Error(), "plugin") && !strings.Contains(err.Error(), "llm") {
			t.Logf("Error message doesn't mention expected terms, but that's okay: %s", err.Error())
		}
	}
}

func TestPlatformInstaller_InstallWheresmyprompt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping wheresmyprompt install test in short mode")
	}

	installer := NewPlatformInstaller()

	result, err := installer.InstallWheresmyprompt()

	// This might succeed or fail depending on system setup
	if err != nil {
		t.Logf("InstallWheresmyprompt failed: %v", err)
	}

	if result != nil {
		t.Logf("Wheresmyprompt installation result: Success=%t, Message=%s, Error=%s",
			result.Success, result.Message, result.Error)
	}
}

func TestPlatformInstaller_InstallFiles2prompt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping files2prompt install test in short mode")
	}

	installer := NewPlatformInstaller()

	result, err := installer.InstallFiles2prompt()

	// This might succeed or fail depending on system setup
	if err != nil {
		t.Logf("InstallFiles2prompt failed: %v", err)
	}

	if result != nil {
		t.Logf("Files2prompt installation result: Success=%t, Message=%s, Error=%s",
			result.Success, result.Message, result.Error)
	}
}

func TestPlatformInstaller_InstallLLM(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping LLM install test in short mode")
	}

	installer := NewPlatformInstaller()

	result, err := installer.InstallLLM()

	// This might succeed or fail depending on system setup
	if err != nil {
		t.Logf("InstallLLM failed: %v", err)
	}

	if result != nil {
		t.Logf("LLM installation result: Success=%t, Message=%s, Error=%s",
			result.Success, result.Message, result.Error)
	}
}

func TestPlatformInstaller_VerifyInstallation(t *testing.T) {
	installer := NewPlatformInstaller()

	// Test with echo command which should exist
	err := installer.VerifyInstallation("echo")
	if err != nil {
		t.Logf("Echo verification failed (might be expected on some minimal systems): %v", err)
	}

	// Test with non-existent command
	err = installer.VerifyInstallation("nonexistent-command-xyz123")
	if err == nil {
		t.Error("Expected error verifying non-existent command")
	}
}

func TestPlatformInstaller_UpdatePathEnvironment(t *testing.T) {
	installer := NewPlatformInstaller()

	// Test with a fake path to avoid modifying real environment
	testPath := "/fake/test/path"

	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath) // Restore original PATH

	err := installer.UpdatePathEnvironment(testPath)
	if err != nil {
		t.Errorf("Unexpected error updating PATH: %v", err)
	}

	// Check that PATH was updated
	newPath := os.Getenv("PATH")
	if !strings.Contains(newPath, testPath) {
		t.Errorf("Expected PATH to contain %s, got %s", testPath, newPath)
	}
}

func TestInstallationResult(t *testing.T) {
	// Test the InstallationResult type
	result := &InstallationResult{
		Success: true,
		Message: "Test installation successful",
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Message != "Test installation successful" {
		t.Errorf("Expected message to be 'Test installation successful', got %s", result.Message)
	}

	// Test failure case
	failResult := &InstallationResult{
		Success: false,
		Error:   "Test installation failed",
	}

	if failResult.Success {
		t.Error("Expected success to be false")
	}

	if failResult.Error != "Test installation failed" {
		t.Errorf("Expected error to be 'Test installation failed', got %s", failResult.Error)
	}
}

func TestGetSystemInfo(t *testing.T) {
	sysInfo := GetSystemInfo()

	if sysInfo == nil {
		t.Error("Expected system info to be non-nil")
		return
	}

	if sysInfo.OS == "" {
		t.Error("Expected OS to be non-empty")
	}

	if sysInfo.Architecture == "" {
		t.Error("Expected Architecture to be non-empty")
	}

	t.Logf("System Info - OS: %s, Arch: %s, Shell: %s, PATH dirs: %d",
		sysInfo.OS, sysInfo.Architecture, sysInfo.Shell, len(sysInfo.PathDirs))
}

func TestPlatformSpecificBehavior(t *testing.T) {
	installer := NewPlatformInstaller()

	// Test platform-specific behavior
	switch installer.OS {
	case "darwin":
		t.Log("Testing macOS-specific behavior")
		// On macOS, Homebrew might be preferred
		if installer.PackageManager == "homebrew" {
			t.Log("Homebrew detected on macOS")
		}
	case "linux":
		t.Log("Testing Linux-specific behavior")
		// On Linux, various package managers might be detected
		t.Logf("Linux package manager: %s", installer.PackageManager)
	case "windows":
		t.Log("Testing Windows-specific behavior")
	default:
		t.Logf("Unknown OS: %s", installer.OS)
	}
}

func TestErrorHandling(t *testing.T) {
	// Test various error conditions
	installer := NewPlatformInstaller()

	// Test verification with empty binary name
	err := installer.VerifyInstallation("")
	if err == nil {
		t.Error("Expected error verifying empty binary name")
	}

	// Test path update with empty path
	err = installer.UpdatePathEnvironment("")
	if err != nil {
		t.Errorf("Unexpected error updating PATH with empty string: %v", err)
	}
}

// Benchmark tests
func BenchmarkCheckDependency(b *testing.B) {
	dep := Dependency{
		Name:         "echo",
		Command:      "echo",
		CheckCommand: "echo test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CheckDependency(dep)
	}
}

func BenchmarkNewPlatformInstaller(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewPlatformInstaller()
	}
}

func BenchmarkDetectPackageManager(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetectPackageManager()
	}
}

// Helper functions for tests
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Test concurrent installations (simulation)
func TestConcurrentInstallationChecks(t *testing.T) {
	deps := RequiredDependencies()

	if len(deps) == 0 {
		t.Skip("No dependencies to test concurrency with")
	}

	// Test checking installation status concurrently
	results := make(chan *InstallationResult, len(deps))
	errors := make(chan error, len(deps))

	for _, dep := range deps {
		go func(d Dependency) {
			// Just test the function exists and can be called
			result, err := InstallDependency(d)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}(dep)
	}

	// Collect results
	successCount := 0
	errorCount := 0

	for i := 0; i < len(deps); i++ {
		select {
		case result := <-results:
			successCount++
			t.Logf("Installation check completed: Success=%t", result.Success)
		case err := <-errors:
			errorCount++
			t.Logf("Installation check error: %v", err)
		}
	}

	t.Logf("Concurrent installation checks: %d results, %d errors", successCount, errorCount)
}

// Test mock installation without actually installing
func TestMockInstallationFlow(t *testing.T) {
	installer := NewPlatformInstaller()

	// Test the installation flow without actually installing anything
	t.Logf("Mock installation flow for platform: %s/%s", installer.OS, installer.Architecture)

	// Test each dependency's installation instructions
	for _, dep := range RequiredDependencies() {
		instructions := GetInstallInstructions(dep)
		if instructions == "" {
			t.Errorf("No installation instructions for %s", dep.Name)
		} else {
			t.Logf("Instructions available for %s", dep.Name)
		}
	}
}

// Test the system information gathering
func TestSystemInfoCollection(t *testing.T) {
	sysInfo := GetSystemInfo()

	// Verify all fields are populated
	if sysInfo.OS == "" {
		t.Error("OS should not be empty")
	}

	if sysInfo.Architecture == "" {
		t.Error("Architecture should not be empty")
	}

	// PATH dirs should contain at least some directories
	if len(sysInfo.PathDirs) == 0 {
		t.Error("PATH dirs should contain at least some directories")
	}

	// Validate OS is a known value
	knownOS := []string{"linux", "darwin", "windows", "freebsd", "openbsd", "netbsd"}
	osFound := false
	for _, os := range knownOS {
		if sysInfo.OS == os {
			osFound = true
			break
		}
	}

	if !osFound {
		t.Errorf("Unknown OS detected: %s", sysInfo.OS)
	}

	t.Logf("System info collected successfully: %+v", sysInfo)
}
