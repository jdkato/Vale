package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
	"github.com/pterm/pterm"

	"github.com/errata-ai/vale/v3/internal/core"
)

const nativeHostName = "sh.vale.native"
const releaseURL = "https://github.com/errata-ai/vale-native/releases/download/%s/vale-native_%s.%s"

var supportedBrowsers = []string{
	"chrome",
	"firefox",
	"opera",
	"chromium",
	"edge",
}

var extensionByBrowser = map[string]string{
	"chrome": "chrome-extension://kfmjcegeklidlnjoechfggipjjjahedj/",
}

var (
	errMissingBrowser = errors.New("missing argument 'browser'")
	errInvalidBrowser = fmt.Errorf("invalid browser; must one of %v", supportedBrowsers)
	errMissingExt     = errors.New("no extension for the given browser")
)

type manifest struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Path              string   `json:"path"`
	Type              string   `json:"type"`
	AllowedExtensions []string `json:"allowed_extensions,omitempty"`
	AllowedOrigins    []string `json:"allowed_origins,omitempty"`
}

// getNativeConfig returns the path to the native host's config file.
//
// NOTE: When the browser (e.g., Chrome) launches the native host, it does
// not have access to the user's shell environment. This is actually why we
// need a config file at all -- to tell the host where to find the Vale
// binary.
//
// The problem is, however, that we can't rely on `XDG_CONFIG_HOME` to be set,
// so we need to use the default value.
func getNativeConfig() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		cfg, notFound := xdg.ConfigFile("vale/native/config.json")
		if notFound != nil {
			return "", notFound
		}
		return cfg, nil
	case "linux":
		path := filepath.Join(home, ".config/vale/native/config.json")
		if err = mkdir(filepath.Dir(path)); err != nil {
			return "", err
		}
		return path, nil
	case "darwin":
		path := filepath.Join(home, "Library/Application Support/vale/native/config.json")
		if err = mkdir(filepath.Dir(path)); err != nil {
			return "", err
		}
		return path, nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func getExecName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func getManifestDirs() (map[string]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	manifests := map[string]string{}
	switch runtime.GOOS {
	case "linux":
		manifests = map[string]string{
			"chrome":   filepath.Join(home, ".config/google-chrome/NativeMessagingHosts"),
			"firefox":  filepath.Join(home, ".mozilla/native-messaging-hosts"),
			"opera":    filepath.Join(home, ".config/google-chrome/NativeMessagingHosts"),
			"chromium": filepath.Join(home, ".config/chromium/NativeMessagingHosts"),
		}
	case "darwin":
		manifests = map[string]string{
			"chrome":   filepath.Join(home, "Library/Application Support/Google/Chrome/NativeMessagingHosts"),
			"firefox":  filepath.Join(home, "Library/Application Support/Mozilla/NativeMessagingHosts"),
			"opera":    filepath.Join(home, "Library/Application Support/Google/Chrome/NativeMessagingHosts"),
			"chromium": filepath.Join(home, "Library/Application Support/Chromium/NativeMessagingHosts"),
			"edge":     filepath.Join(home, "Library/Application Support/Microsoft Edge/NativeMessagingHosts"),
		}
	}

	return manifests, nil
}

func getLocation(browser string) (map[string]string, error) {
	cfg, err := getNativeConfig()
	if err != nil {
		return nil, err
	}

	bin := filepath.Dir(cfg)
	if runtime.GOOS == "windows" {
		return map[string]string{
			"appDir":      bin,
			"manifestDir": "",
		}, nil
	}

	manifestDirs, err := getManifestDirs()
	if err != nil {
		return nil, err
	}

	manifest := ""
	if found, ok := manifestDirs[browser]; ok {
		manifest = found
	}

	return map[string]string{
		"appDir":      bin,
		"manifestDir": manifest,
	}, nil
}

func writeNativeConfig() (string, error) {
	cfgFile, err := getNativeConfig()
	if err != nil {
		return "", err
	}

	exe, err := exec.LookPath("vale")
	if err != nil {
		return "", err
	}

	cfg := map[string]string{
		"path": exe,
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}

	return cfgFile, os.WriteFile(cfgFile, jsonCfg, 0600)
}

func installNativeHostUnix(manifestData []byte, manifestFile string) error {
	err := os.WriteFile(manifestFile, manifestData, 0600)
	if err != nil {
		return err
	}
	return nil
}

func installNativeHostWindows(manifestData []byte, manifestFile, browser string) error {
	cfg, err := getNativeConfig()
	if err != nil {
		return err
	}

	manifestDir := filepath.Join(filepath.Dir(cfg), "manifest", browser)

	err = os.MkdirAll(manifestDir, os.ModePerm)
	if err != nil {
		return err
	}
	subdir := filepath.Join(manifestDir, manifestFile)

	err = os.WriteFile(subdir, manifestData, 0600)
	if err != nil {
		return err
	}

	err = setManifestRegistry(browser, subdir)
	if err != nil {
		return err
	}

	return nil
}

func getLatestHostRelease() (string, error) {
	resp, err := fetchJSON("https://api.github.com/repos/errata-ai/vale-native/releases/latest")
	if err != nil {
		return "", err
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	err = json.Unmarshal(resp, &release)
	if err != nil {
		return "", err
	}

	return release.TagName, nil
}

func hostDownloadURL() (string, error) {
	hostVersion, err := getLatestHostRelease()
	if err != nil {
		return "", err
	}
	name := platformAndArch()
	return fmt.Sprintf(releaseURL, hostVersion, name, "zip"), nil
}

func installHost(manifestJSON []byte, manifestFile, browser string) error {
	switch runtime.GOOS {
	case "linux", "darwin":
		return installNativeHostUnix(manifestJSON, manifestFile)
	case "windows":
		return installNativeHostWindows(manifestJSON, manifestFile, browser)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func installNativeHost(args []string, _ *core.CLIFlags) error { //nolint:funlen
	if len(args) != 1 {
		return core.NewE100("host-install", errMissingBrowser)
	}

	browser := args[0]
	if !core.StringInSlice(browser, supportedBrowsers) {
		return core.NewE100("host-install", errInvalidBrowser)
	}

	steps := []string{"writing config", "fetching binary", "installing host"}
	p, _ := pterm.DefaultProgressbar.WithTotal(len(steps)).WithTitle("Installing host").Start()

	p.UpdateTitle(steps[0])
	cfgFile, err := writeNativeConfig()
	if err != nil {
		return progressError("host-install", err, p)
	}
	pterm.Success.Println(fmt.Sprintf("wrote '%s'", cfgFile))
	p.Increment()

	locations, err := getLocation(browser)
	if err != nil {
		return progressError("host-install", err, p)
	}

	hostURL, err := hostDownloadURL()
	if err != nil {
		return progressError("host-install", err, p)
	}
	exeName := getExecName("vale-native")

	oldInstall := []string{exeName, "LICENSE", "README.md"}
	for _, file := range oldInstall {
		fp := filepath.Join(locations["appDir"], file)
		if core.FileExists(fp) {
			err = os.Remove(fp)
			if err != nil {
				return progressError("host-install", err, p)
			}
		}
	}

	p.UpdateTitle(steps[1])
	err = fetch(hostURL, locations["appDir"])
	if err != nil {
		return progressError("host-install", err, p)
	}
	pterm.Success.Println(fmt.Sprintf("fetched '%s'", hostURL))
	p.Increment()

	manifestData := manifest{
		Name:        nativeHostName,
		Description: "A native messaging for the Vale CLI.",
		Type:        "stdio",
		Path:        filepath.Join(locations["appDir"], exeName),
	}

	manifestFile := filepath.Join(locations["manifestDir"], manifestData.Name+".json")

	extension, found := extensionByBrowser[browser]
	if !found {
		return progressError("host-install", errMissingExt, p)
	}
	allowed := []string{extension}

	devID := fmt.Sprintf("VALE_DEV_%s_ID", strings.ToUpper(browser))
	if id, set := os.LookupEnv(devID); set {
		allowed = append(allowed, id)
	}

	if browser == "firefox" {
		manifestData.AllowedExtensions = allowed
	} else {
		manifestData.AllowedOrigins = allowed
	}

	manifestJSON, err := json.MarshalIndent(manifestData, "", "  ")
	if err != nil {
		return progressError("host-install", err, p)
	}

	p.UpdateTitle(steps[2])
	err = installHost(manifestJSON, manifestFile, browser)
	if err != nil {
		return progressError("host-install", err, p)
	}
	pterm.Success.Println(fmt.Sprintf("installed '%s'", manifestFile))
	p.Increment()

	return nil
}

func uninstallNativeHost(args []string, _ *core.CLIFlags) error {
	if len(args) != 1 {
		return core.NewE100("host-uninstall", errMissingBrowser)
	}

	browser := args[0]
	if !core.StringInSlice(browser, supportedBrowsers) {
		return core.NewE100("host-uninstall", errInvalidBrowser)
	}

	steps := []string{"removing files", "uninstalling host"}
	p, _ := pterm.DefaultProgressbar.WithTotal(len(steps)).WithTitle("Uninstalling host").Start()

	locations, err := getLocation(browser)
	if err != nil {
		return progressError("host-uninstall", err, p)
	}
	p.UpdateTitle(steps[0])

	exeName := getExecName("vale-native")
	for _, file := range []string{"config.json", exeName, "LICENSE", "README.md", "host.log"} {
		fp := filepath.Join(locations["appDir"], file)
		if core.FileExists(fp) {
			err = os.Remove(filepath.Join(locations["appDir"], file))
			if err != nil {
				return progressError("host-uninstall", err, p)
			}
		}
	}
	pterm.Success.Println(steps[0])
	p.Increment()

	p.UpdateTitle(steps[1])
	manifestFile := filepath.Join(locations["manifestDir"], nativeHostName+".json")

	if core.FileExists(manifestFile) {
		err = os.Remove(manifestFile)
		if err != nil {
			return progressError("host-uninstall", err, p)
		}
	}

	err = unsetManifestRegistry(browser)
	if err != nil {
		return progressError("host-uninstall", err, p)
	}

	pterm.Success.Println(steps[1])
	p.Increment()

	return nil
}
