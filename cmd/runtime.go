package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var runtimeCommandsOnce sync.Once

func registerRuntimeCommands() {
	runtimeCommandsOnce.Do(func() {
		RootCmd.AddCommand(newTopLevelSecretAlias("get", secretsGetCmd))
		RootCmd.AddCommand(newTopLevelSecretAlias("set", secretsSetCmd))
		RootCmd.AddCommand(newTopLevelSecretAlias("ls", secretsLsCmd))
		RootCmd.AddCommand(newTopLevelSecretAlias("rm", secretsRmCmd))
		RootCmd.AddCommand(newTopLevelSecretAlias("ensure", secretsEnsureCmd))
		RootCmd.AddCommand(newHiddenVersionCmd())
		RootCmd.AddCommand(newUpgradeCmd())
	})
}

func newTopLevelSecretAlias(name string, base *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   base.Use,
		Short: base.Short,
		Long:  base.Long,
		Args:  base.Args,
		Run:   base.Run,
	}
	cmd.Flags().AddFlagSet(base.LocalFlags())
	cmd.PersistentFlags().AddFlagSet(base.PersistentFlags())
	cmd.SilenceErrors = base.SilenceErrors
	cmd.SilenceUsage = base.SilenceUsage
	cmd.Example = base.Example
	cmd.Hidden = false
	return cmd
}

func newHiddenVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "version",
		Short:  "Print the kpv version",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), VERSION)
		},
	}
}

type githubRelease struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func newUpgradeCmd() *cobra.Command {
	var preRelease bool

	cmd := &cobra.Command{
		Use:   "upgrade [version]",
		Short: "Upgrade kpv in place",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := "latest"
			if len(args) > 0 && args[0] != "" {
				version = args[0]
			}

			execPath, err := os.Executable()
			if err != nil {
				return err
			}

			if err := upgradeBinary(execPath, version, preRelease); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "kpv upgraded successfully")
			return err
		},
	}

	cmd.Flags().BoolVar(&preRelease, "prerelease", false, "Allow prerelease versions when upgrading")
	cmd.Flags().BoolVar(&preRelease, "pre-release", false, "Allow prerelease versions when upgrading")
	return cmd
}

func upgradeBinary(destPath, version string, preRelease bool) error {
	release, err := fetchRelease(version, preRelease)
	if err != nil {
		return err
	}

	assetName := fmt.Sprintf("kpv-%s-%s-v%s", runtime.GOOS, runtime.GOARCH, strings.TrimPrefix(release.TagName, "v"))
	archiveExt := "tar.gz"
	if runtime.GOOS == "windows" {
		archiveExt = "zip"
	}

	assetURL := ""
	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.Name, assetName) && strings.HasSuffix(asset.Name, "."+archiveExt) {
			assetURL = asset.URL
			break
		}
	}
	if assetURL == "" {
		return fmt.Errorf("no release asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	tmpDir, err := os.MkdirTemp("", "kpv-upgrade-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, "kpv."+archiveExt)
	if err := downloadToFile(assetURL, archivePath); err != nil {
		return err
	}

	if err := extractUpgradeArchive(archivePath, tmpDir); err != nil {
		return err
	}

	binaryPath := filepath.Join(tmpDir, binaryName())
	return replaceUpgradeBinary(binaryPath, destPath)
}

func fetchRelease(version string, preRelease bool) (*githubRelease, error) {
	url := releaseAPIURL("frostyeti", "kpv", version, preRelease)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "kpv-upgrade")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch release: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	if version == "latest" && preRelease {
		var releases []githubRelease
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, err
		}
		for _, release := range releases {
			if release.Prerelease {
				return &release, nil
			}
		}
		if len(releases) == 0 {
			return nil, fmt.Errorf("no releases returned")
		}
		return &releases[0], nil
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	if release.TagName == "" {
		return nil, fmt.Errorf("empty release tag")
	}
	return &release, nil
}

func releaseAPIURL(owner, repo, version string, preRelease bool) string {
	if version == "latest" {
		if preRelease {
			return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)
		}
		return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	}
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/v%s", owner, repo, version)
}

func downloadToFile(url, dest string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "kpv-upgrade")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func extractUpgradeArchive(archive, dir string) error {
	if strings.HasSuffix(archive, ".zip") {
		zr, err := zip.OpenReader(archive)
		if err != nil {
			return err
		}
		defer zr.Close()
		for _, file := range zr.File {
			if filepath.Base(file.Name) != binaryName() {
				continue
			}
			rc, err := file.Open()
			if err != nil {
				return err
			}
			outPath := filepath.Join(dir, binaryName())
			out, err := os.Create(outPath)
			if err != nil {
				rc.Close()
				return err
			}
			if _, err := io.Copy(out, rc); err != nil {
				rc.Close()
				out.Close()
				return err
			}
			rc.Close()
			return out.Close()
		}
		return fmt.Errorf("binary not found in archive")
	}

	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(hdr.Name) != binaryName() {
			continue
		}

		outPath := filepath.Join(dir, binaryName())
		out, err := os.Create(outPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	}

	return fmt.Errorf("binary not found in archive")
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "kpv.exe"
	}
	return "kpv"
}

func replaceUpgradeBinary(newBinary, destPath string) error {
	if runtime.GOOS == "windows" {
		return os.Rename(newBinary, destPath)
	}
	if err := os.Rename(newBinary, destPath); err != nil {
		return err
	}
	return os.Chmod(destPath, 0755)
}
