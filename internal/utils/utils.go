package utils

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/99designs/keyring"
	"github.com/frostyeti/go/secrets"
	"github.com/frostyeti/kpv/internal/keepass"

	"github.com/frostyeti/kpv/internal/kvconf"
	"github.com/spf13/cobra"
)

var cachedConf *kvconf.Config

func GetConfileFile() (string, error) {
	configFile := os.Getenv("KPV_CONFIG_FILE")
	if configFile != "" {
		return configFile, nil
	}

	configDir := os.Getenv("KPV_CONFIG_DIR")
	if configDir != "" {
		return filepath.Join(configDir, "kpv.kvc"), nil
	}

	home, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.New("unable to determine user home directory for config file")
		}

		if runtime.GOOS == "windows" {
			configDir = filepath.Join(home, "AppData", "Roaming", "kpv")
		} else {
			configDir = filepath.Join(home, ".config", "kpv")
		}

		return filepath.Join(configDir, "kpv.kvc"), nil
	}

	return filepath.Join(home, "kpv", "kpv.kvc"), nil
}

func GetConfig() (*kvconf.Config, error) {
	if cachedConf != nil {
		return cachedConf, nil
	}

	configFile, err := GetConfileFile()
	if err != nil {
		return nil, err
	}

	conf := kvconf.NewConfig()

	err = conf.Load(configFile)
	if err != nil {
		return nil, err
	}

	cachedConf = conf
	return conf, nil
}

type ResolvedPath struct {
	Path string
}

func Err(msg string) {
	os.Stderr.WriteString("\x1b[31m[error]:\x1b[0m ")
	os.Stderr.WriteString(msg)
	os.Stderr.WriteString("\n")
}

func Errf(format string, args ...interface{}) {
	os.Stderr.WriteString("\x1b[31m[error]:\x1b[0m ")
	os.Stderr.WriteString(fmt.Sprintf(format, args...))
	os.Stderr.WriteString("\n")
}

func Fail(msg string) {
	os.Stderr.WriteString("\x1b[31m[error]:\x1b[0m ")
	os.Stderr.WriteString(msg)
	os.Stderr.WriteString("\n")
	os.Exit(1)
}

func Failf(format string, args ...interface{}) {
	os.Stderr.WriteString("\x1b[31m[error]:\x1b[0m ")
	os.Stderr.WriteString(fmt.Sprintf(format, args...))
	os.Stderr.WriteString("\n")
	os.Exit(1)
}

func Warn(msg string) {
	os.Stderr.WriteString("\x1b[33m[warn]:\x1b[0m ")
	os.Stderr.WriteString(msg)
	os.Stderr.WriteString("\n")
}

func Warnf(format string, args ...interface{}) {
	os.Stderr.WriteString("\x1b[33m[warning]:\x1b[0m ")
	os.Stderr.WriteString(fmt.Sprintf(format, args...))
	os.Stderr.WriteString("\n")
}

func Ok(msg string) {
	os.Stderr.WriteString("\x1b[32m[ok]:\x1b[0m ")
	os.Stderr.WriteString(msg)
	os.Stderr.WriteString("\n")
}

func Okf(format string, args ...interface{}) {
	os.Stderr.WriteString("\x1b[32m[ok]:\x1b[0m ")
	os.Stderr.WriteString(fmt.Sprintf(format, args...))
	os.Stderr.WriteString("\n")
}

func SetAlias(name, path string) error {
	conf, err := GetConfig()
	if err != nil {
		return err
	}

	aliases, _ := conf.Get("aliases")
	lines := []string{}
	if aliases != "" {
		lines = strings.Split(aliases, "\n")
	}

	updated := false
	for i, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			existingName := strings.TrimSpace(parts[0])
			if strings.EqualFold(existingName, name) {
				lines[i] = fmt.Sprintf("%s=%s", name, path)
				updated = true
				break
			}
		}
	}

	if !updated {
		lines = append(lines, fmt.Sprintf("%s=%s", name, path))
	}

	conf.Set("aliases", strings.Join(lines, "\n"))
	return conf.Save()
}

func ResolveVaultPath(vault string) (ResolvedPath, error) {
	if vault == "" {
		vault = "default"
	}

	if vault != "" {
		if strings.HasPrefix(vault, "file:///") || strings.HasPrefix(vault, "kpv:///") {
			parsed, err := url.Parse(vault)
			if err != nil {
				return ResolvedPath{}, err
			}
			return ResolvedPath{Path: parsed.Path}, nil
		}

		conf, err := GetConfig()
		if err == nil {
			if vault == "default" {
				defaultPath, ok := conf.Get("defaults.path")
				if ok {
					return ResolvedPath{Path: defaultPath}, nil
				}

				defaultPath = GetDefaultVaultPath("default.kdbx")
				return ResolvedPath{Path: defaultPath}, nil
			}

			aliases, ok := conf.Get("aliases")
			if ok {
				lines := strings.Split(aliases, "\n")
				for _, line := range lines {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						name := strings.TrimSpace(parts[0])
						path := strings.TrimSpace(parts[1])
						if strings.EqualFold(name, vault) {
							return ResolvedPath{Path: path}, nil
						}
					}
				}
			}
		} else {
			if vault == "default" {
				defaultPath := GetDefaultVaultPath("default.kdbx")
				return ResolvedPath{Path: defaultPath}, nil
			}
		}
	}

	// vault is provided
	// Check if it's an absolute path
	if filepath.IsAbs(vault) {
		return ResolvedPath{Path: vault}, nil
	}

	// Check if it's a relative path or exists in current directory
	if strings.Contains(vault, string(filepath.Separator)) || strings.Contains(vault, "/") {
		absPath, err := filepath.Abs(vault)
		if err != nil {
			return ResolvedPath{}, err
		}
		return ResolvedPath{Path: absPath}, nil
	}

	// Check current directory
	currentDirPath := vault
	if !strings.HasSuffix(vault, ".kdbx") {
		currentDirPath = vault + ".kdbx"
	}
	if _, err := os.Stat(currentDirPath); err == nil {
		absPath, err := filepath.Abs(currentDirPath)
		if err != nil {
			return ResolvedPath{}, err
		}
		return ResolvedPath{Path: absPath}, nil
	}

	// Check in ~/.local/share/kpv or %LOCALAPPDATA%/kpv
	vaultName := vault
	if !strings.HasSuffix(vaultName, ".kdbx") {
		vaultName = vaultName + ".kdbx"
	}
	defaultPath := GetDefaultVaultPath(vaultName)
	return ResolvedPath{Path: defaultPath}, nil
}

func GetDefaultVaultPath(filename string) string {
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		return filepath.Join(localAppData, "kpv", filename)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".local", "share", "kpv", filename)
}

func GetPassword(cmd *cobra.Command, vaultPath string) (*string, error) {
	// First check command line flag
	password, _ := cmd.Flags().GetString("password")
	if password != "" {
		return &password, nil
	}

	// Then check password file flag
	passwordFile, _ := cmd.Flags().GetString("password-file")
	if passwordFile != "" {
		data, err := os.ReadFile(passwordFile)
		if err != nil {
			return nil, err
		}
		pwd := strings.TrimSpace(string(data))
		return &pwd, nil
	}

	// Try to get from OS keyring
	kr, err := OpenKeyring()
	if err == nil {
		item, err := kr.Get("kpv:///" + vaultPath)
		if err == nil {
			pwd := string(item.Data)
			return &pwd, nil
		}
	}

	// Finally, check for .key file in default location
	// Use vault base name for key file
	baseName := filepath.Base(vaultPath)
	keyFileName := baseName[:len(baseName)-len(filepath.Ext(baseName))] + ".key"
	keyFilePath := filepath.Join(GetDefaultVaultPath(""), keyFileName)

	if data, err := os.ReadFile(keyFilePath); err == nil {
		pwd := strings.TrimSpace(string(data))
		if pwd != "" {
			return &pwd, nil
		}
	}

	return nil, errors.New("password not provided via --password flag, --password-file flag, KPV_PASSWORD/KPV_PASSWORD_FILE environment variables, OS keyring, or .key file")
}

func OpenKeyring() (keyring.Keyring, error) {
	serviceName := "kpv"
	keychain := "login"
	libsecret := "login"

	conf, err := GetConfig()
	if err == nil {
		if v, ok := conf.Get("libsecret.collection"); ok && v != "" {
			libsecret = v
		}
		if v, ok := conf.Get("keychain.name"); ok && v != "" {
			keychain = v
		}
		if v, ok := conf.Get("service.name"); ok && v != "" {
			serviceName = v
		}
	}

	kr, err := keyring.Open(keyring.Config{
		ServiceName:             serviceName,
		LibSecretCollectionName: libsecret,
		KeychainName:            keychain,
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
		},
	})
	return kr, err
}

func OpenKeePass(cmd *cobra.Command) (*keepass.Kdbx, string, error) {
	vault, _ := cmd.Flags().GetString("vault")

	resolved, err := ResolveVaultPath(vault)
	if err != nil {
		return nil, "", err
	}

	vaultPath := resolved.Path

	password, err := GetPassword(cmd, vaultPath)
	if err != nil {
		return nil, "", err
	}

	options := keepass.KdbxOptions{
		Path:      vaultPath,
		Secret:    password,
		Create:    true,
		CreateDir: true,
	}

	kdbx, err := keepass.Open(options)
	if err != nil {
		return nil, "", err
	}

	return kdbx, vaultPath, nil
}

// GenerateSecretWithOptions generates a secret using the provided options
// This function is shared across import and sync commands
func GenerateSecretWithOptions(size int, noUpper, noLower, noDigits, noSpecial bool, special, chars string) (string, error) {
	builder := secrets.NewOptionsBuilder()
	builder.WithSize(int16(size))
	builder.WithRetries(100)

	if chars != "" {
		// If chars is specified, use only those characters
		builder.WithChars(chars)
	} else {
		// Otherwise, build character set from flags
		builder.WithUpper(!noUpper)
		builder.WithLower(!noLower)
		builder.WithDigits(!noDigits)

		if noSpecial {
			builder.WithNoSymbols()
		} else if special != "" {
			builder.WithSymbols(special)
		} else {
			// Default special characters
			builder.WithSymbols("@_-{}|#!`~:^")
		}
	}

	opts := builder.Build()
	return opts.Generate()
}
