package cmd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	repoName string
	version  string
	filePath string
	dryRun   bool
)

var rootCmd = &cobra.Command{
	Use:   "brewup",
	Short: "Update Homebrew formula with new version and checksums",
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateFormula()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&repoName, "repo", "r", "", "Repository name (e.g., sbomasm)")
	rootCmd.Flags().StringVarP(&version, "version", "v", "", "Version tag (e.g., v1.0.5)")
	rootCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to Homebrew formula file (e.g., sbomasm.rb)")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without modifying the file")
	rootCmd.MarkFlagRequired("repo")
	rootCmd.MarkFlagRequired("version")
	rootCmd.MarkFlagRequired("file")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func updateFormula() error {
	// Validate inputs
	if !strings.HasPrefix(version, "v") {
		return fmt.Errorf("version must start with 'v' (e.g., v1.0.5)")
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("formula file does not exist: %s", filePath)
	}

	// Read the formula file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read formula file: %w", err)
	}
	originalContent := string(content)

	// Update version
	versionRegex := regexp.MustCompile(`version\s+"v\d+\.\d+\.\d+"`)
	newVersion := fmt.Sprintf(`version "%s"`, version)
	updatedContent := versionRegex.ReplaceAllString(originalContent, newVersion)

	// Define platforms and their binary names
	platforms := []struct {
		os   string
		arch string
	}{
		{"darwin", "arm64"},
		{"darwin", "amd64"},
		{"linux", "arm64"},
		{"linux", "amd64"},
	}

	// Update URLs and checksums for each platform
	for _, p := range platforms {
		binaryName := fmt.Sprintf("%s-%s-%s", repoName, p.os, p.arch)
		newURL := fmt.Sprintf("https://github.com/interlynk-io/%s/releases/download/%s/%s", repoName, version, binaryName)

		// Download binary and calculate checksum
		checksum, err := calculateChecksum(newURL)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum for %s: %w", binaryName, err)
		}

		// Update URL
		urlRegex := regexp.MustCompile(fmt.Sprintf(`url "https://github\.com/interlynk-io/%s/releases/download/v\d+\.\d+\.\d+/%s",\s*:using\s*=>\s*:nounzip`, regexp.QuoteMeta(repoName), regexp.QuoteMeta(binaryName)))
		updatedContent = urlRegex.ReplaceAllString(updatedContent, fmt.Sprintf(`url "%s", :using => :nounzip`, newURL))

		// Update checksum
		checksumRegex := regexp.MustCompile(fmt.Sprintf(`(url "%s",\s*:using\s*=>\s*:nounzip\n\s*sha256 ")[0-9a-f]{64}"`, regexp.QuoteMeta(newURL)))
		updatedContent = checksumRegex.ReplaceAllString(updatedContent, fmt.Sprintf(`$1"%s"`, checksum))
	}

	// Print changes (dry-run or log)
	fmt.Printf("Changes to %s:\n", filePath)
	fmt.Printf("Version: %s -> %s\n", versionRegex.FindString(originalContent), newVersion)
	for _, p := range platforms {
		binaryName := fmt.Sprintf("%s-%s-%s", repoName, p.os, p.arch)
		// oldURL := fmt.Sprintf("https://github.com/interlynk-io/%s/releases/download/v\\d+\\.\\d+\\.\\d+/%s", repoName, binaryName)
		newURL := fmt.Sprintf("https://github.com/interlynk-io/%s/releases/download/%s/%s", repoName, version, binaryName)

		// Extract old checksum
		oldChecksumRegex := regexp.MustCompile(fmt.Sprintf(`(url "https://github\.com/interlynk-io/%s/releases/download/v\d+\.\d+\.\d+/%s",\s*:using\s*=>\s*:nounzip\n\s*sha256 ")[0-9a-f]{64}"`, regexp.QuoteMeta(repoName), regexp.QuoteMeta(binaryName)))
		oldChecksumMatch := oldChecksumRegex.FindString(originalContent)
		var oldChecksum string
		if oldChecksumMatch != "" {
			oldChecksum = regexp.MustCompile(`[0-9a-f]{64}`).FindString(oldChecksumMatch)
		}

		// Extract new checksum
		newChecksumRegex := regexp.MustCompile(fmt.Sprintf(`(url "%s",\s*:using\s*=>\s*:nounzip\n\s*sha256 ")[0-9a-f]{64}"`, regexp.QuoteMeta(newURL)))
		newChecksumMatch := newChecksumRegex.FindString(updatedContent)
		var newChecksum string
		if newChecksumMatch != "" {
			newChecksum = regexp.MustCompile(`[0-9a-f]{64}`).FindString(newChecksumMatch)
		}

		fmt.Printf("Checksum (%s-%s): %s -> %s\n", p.os, p.arch, oldChecksum, newChecksum)
	}

	// Write changes (unless dry-run)
	if dryRun {
		fmt.Println("Dry-run mode: No changes written to file")
		fmt.Println("Updated content preview:")
		fmt.Println(updatedContent)
		return nil
	}

	if err := os.WriteFile(filePath, []byte(updatedContent), 0o644); err != nil {
		return fmt.Errorf("failed to write updated formula file: %w", err)
	}

	fmt.Printf("Successfully updated %s\n", filePath)
	return nil
}

func calculateChecksum(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download %s: status %s", url, resp.Status)
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, resp.Body); err != nil {
		return "", fmt.Errorf("failed to compute checksum: %w", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
