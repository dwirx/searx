package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	latestReleaseURL = "https://api.github.com/repos/lightpanda-io/browser/releases/latest"
	downloadURLTmpl  = "https://github.com/lightpanda-io/browser/releases/download/%s/lightpanda-%s-linux"
)

var semverPattern = regexp.MustCompile(`v?(\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?)`)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func configuredLightpandaPath() string {
	if custom := strings.TrimSpace(os.Getenv("SEARX_LIGHTPANDA_PATH")); custom != "" {
		return custom
	}

	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "./lightpanda"
	}

	return filepath.Join(home, ".local", "share", "searx", "lightpanda")
}

func resolveLightpandaBinaryPath() (string, bool) {
	configured := configuredLightpandaPath()
	if info, err := os.Stat(configured); err == nil && !info.IsDir() {
		return configured, true
	}

	if fromPath, err := exec.LookPath("lightpanda"); err == nil {
		return fromPath, true
	}

	return configured, false
}

func LightpandaBinaryPath() (string, error) {
	if path, ok := resolveLightpandaBinaryPath(); ok {
		return path, nil
	}

	return "", fmt.Errorf("lightpanda not installed; run `search setup`")
}

func normalizeLightpandaVersion(raw string) (string, bool) {
	match := semverPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if len(match) < 2 {
		return "", false
	}

	return match[1], true
}

func sameLightpandaVersion(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" || b == "" {
		return false
	}
	if strings.EqualFold(a, b) {
		return true
	}

	na, oka := normalizeLightpandaVersion(a)
	nb, okb := normalizeLightpandaVersion(b)
	return oka && okb && na == nb
}

func isLightpandaUpToDate(localVersion, latestTag, recordedTag string) bool {
	if sameLightpandaVersion(recordedTag, latestTag) {
		return true
	}
	return sameLightpandaVersion(localVersion, latestTag)
}

func shouldAutoUpdateLightpanda(localVersion, latestTag, recordedTag string) bool {
	if localVersion == "Not installed" {
		return true
	}
	if strings.Contains(localVersion, "Error") {
		return true
	}
	if isLightpandaUpToDate(localVersion, latestTag, recordedTag) {
		return false
	}
	if _, ok := normalizeLightpandaVersion(localVersion); !ok {
		// Unknown local version (often a commit hash). Don't auto-update on normal reads.
		return false
	}
	return true
}

func lightpandaTagPath(binaryPath string) string {
	return binaryPath + ".tag"
}

func readRecordedLightpandaTag(binaryPath string) string {
	tagBytes, err := os.ReadFile(lightpandaTagPath(binaryPath))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(tagBytes))
}

func writeRecordedLightpandaTag(binaryPath, tag string) error {
	return os.WriteFile(lightpandaTagPath(binaryPath), []byte(strings.TrimSpace(tag)+"\n"), 0644)
}

func GetLocalLightpandaVersion() string {
	binaryPath, ok := resolveLightpandaBinaryPath()
	if !ok {
		return "Not installed"
	}

	cmd := exec.Command(binaryPath, "version")
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return "Unknown (Error running binary)"
	}

	version := strings.TrimSpace(out.String())
	if version == "" {
		version = strings.TrimSpace(errOut.String())
	}
	if version == "" {
		return "Unknown (No version output)"
	}
	return version
}

func GetLatestLightpandaVersion() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(latestReleaseURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

func EnsureLightpanda() error {
	if runtime.GOOS != "linux" {
		fmt.Printf("Lightpanda auto-setup is only supported on Linux (current: %s). Skipping setup.\n", runtime.GOOS)
		return nil
	}

	local := GetLocalLightpandaVersion()
	recordedTag := readRecordedLightpandaTag(configuredLightpandaPath())

	latest, err := GetLatestLightpandaVersion()
	if err != nil {
		if local != "Not installed" && !strings.Contains(local, "Error") {
			fmt.Printf("Warning: Could not check latest Lightpanda version (%v). Using installed version: %s\n", err, local)
			return nil
		}
		fmt.Printf("Warning: Could not check latest version (%v). Proceeding with nightly if missing.\n", err)
		latest = "nightly"
	}

	if !shouldAutoUpdateLightpanda(local, latest, recordedTag) {
		if isLightpandaUpToDate(local, latest, recordedTag) {
			fmt.Printf("Lightpanda is already up to date (%s).\n", latest)
		} else if local != "Not installed" {
			fmt.Printf("Lightpanda detected (%s) but version cannot be verified. Skipping auto-update. Run `search update` to force update.\n", local)
		}
		return nil
	}

	if local != "Not installed" && !strings.Contains(local, "Error") {
		fmt.Printf("Updating Lightpanda from %s to %s...\n", local, latest)
	}

	return downloadLightpanda(latest)
}

func UpdateLightpanda() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("lightpanda update is only supported on Linux (current: %s)", runtime.GOOS)
	}

	latest, err := GetLatestLightpandaVersion()
	if err != nil {
		return fmt.Errorf("could not check latest version: %v", err)
	}

	local := GetLocalLightpandaVersion()
	recordedTag := readRecordedLightpandaTag(configuredLightpandaPath())

	if local != "Not installed" && isLightpandaUpToDate(local, latest, recordedTag) {
		fmt.Printf("Lightpanda is already up to date (%s).\n", latest)
		return nil
	}

	fmt.Printf("Updating Lightpanda from %s to %s...\n", local, latest)
	return downloadLightpanda(latest)
}

func downloadLightpanda(version string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("automatic Lightpanda download is only supported on Linux")
	}

	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	default:
		return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	url := fmt.Sprintf(downloadURLTmpl, version, arch)
	if version == "nightly" {
		url = fmt.Sprintf("https://github.com/lightpanda-io/browser/releases/download/nightly/lightpanda-%s-linux", arch)
	}

	binaryPath := configuredLightpandaPath()
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0755); err != nil {
		return err
	}

	fmt.Printf("Downloading Lightpanda %s...\n", version)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: %s", resp.Status)
	}

	out, err := os.Create(binaryPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Progress bar implementation
	size := resp.ContentLength
	buffer := make([]byte, 32*1024)
	var downloaded int64

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := out.Write(buffer[:n])
			if writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)
			printProgress(downloaded, size)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	fmt.Println("\n[✔] Download complete!")
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return err
	}
	if err := writeRecordedLightpandaTag(binaryPath, version); err != nil {
		return err
	}

	fmt.Printf("[✔] Lightpanda installed at: %s\n", binaryPath)
	return nil
}

func printProgress(downloaded, total int64) {
	if total <= 0 {
		fmt.Printf("\rDownloading... %d bytes", downloaded)
		return
	}
	percent := float64(downloaded) / float64(total) * 100
	bars := int(percent / 2)
	fmt.Printf("\r[%s%s] %.2f%% (%d/%d bytes)",
		strings.Repeat("=", bars),
		strings.Repeat(" ", 50-bars),
		percent, downloaded, total)
}
