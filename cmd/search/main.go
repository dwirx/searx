package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"searx-cli/internal/engine"
	"searx-cli/internal/reader"
	"searx-cli/internal/ui"
	"searx-cli/internal/util"
	"strings"
)

var version = "dev"
var installScriptURL = "https://github.com/dwirx/searx/releases/latest/download/install.sh"

type subcommandOptions struct {
	LightpandaOnly bool
	KeepLightpanda bool
}

func getDefaultEngine() string {
	return "ddg"
}

func shouldFallback(primary string, err error) bool {
	if err == nil {
		return false
	}

	switch strings.ToLower(primary) {
	case "ddg", "searx":
	default:
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "status code 202") ||
		strings.Contains(msg, "status code 403") ||
		strings.Contains(msg, "status code 429")
}

func fallbackEngineNames(primary string) []string {
	switch strings.ToLower(primary) {
	case "ddg":
		return []string{"mojeek", "google", "brave"}
	case "searx":
		return []string{"ddg", "mojeek", "google"}
	default:
		return nil
	}
}

func buildEngine(name, searxInstance, hnCategory string) (engine.SearchEngine, bool) {
	switch strings.ToLower(name) {
	case "google":
		return &engine.GoogleEngine{}, true
	case "ddg":
		return &engine.DuckDuckGoEngine{}, true
	case "brave":
		return &engine.BraveEngine{}, true
	case "mojeek":
		return &engine.MojeekEngine{}, true
	case "hn":
		return &engine.HackerNewsEngine{Category: hnCategory}, true
	case "searx":
		return &engine.SearxEngine{InstanceURL: searxInstance}, true
	default:
		return nil, false
	}
}

func installerActionArgs(action string, keepLightpanda bool) ([]string, error) {
	switch action {
	case "update":
		return []string{"--update"}, nil
	case "uninstall":
		args := []string{"--uninstall"}
		if keepLightpanda {
			args = append(args, "--keep-lightpanda")
		}
		return args, nil
	default:
		return nil, errors.New("unsupported installer action")
	}
}

func installerShellCommand(action string, keepLightpanda bool) (string, error) {
	args, err := installerActionArgs(action, keepLightpanda)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("curl -sSL %s | bash -s -- %s", installScriptURL, strings.Join(args, " ")), nil
}

func runInstallerAction(action string, keepLightpanda bool) error {
	cmdStr, err := installerShellCommand(action, keepLightpanda)
	if err != nil {
		return err
	}

	cmd := exec.Command("bash", "-lc", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "SEARX_SKIP_SETUP=1")
	return cmd.Run()
}

func parseSubcommandOptions(command string, args []string) (subcommandOptions, error) {
	opts := subcommandOptions{}
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	switch command {
	case "update":
		lightpandaOnly := fs.Bool("lightpanda-only", false, "With update command: only check/update Lightpanda")
		if err := fs.Parse(args); err != nil {
			return opts, err
		}
		opts.LightpandaOnly = *lightpandaOnly
	case "uninstall":
		keepLightpanda := fs.Bool("keep-lightpanda", false, "With uninstall command: keep Lightpanda files")
		if err := fs.Parse(args); err != nil {
			return opts, err
		}
		opts.KeepLightpanda = *keepLightpanda
	default:
		return opts, nil
	}

	if fs.NArg() > 0 {
		return opts, fmt.Errorf("unexpected arguments for %s: %s", command, strings.Join(fs.Args(), " "))
	}

	return opts, nil
}

var defaultSearxInstances = []string{
	"https://searx.be",
	"https://paulgo.io",
	"https://searx.org",
}

func main() {
	// Custom usage/help
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: search [options] [query]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  setup         Download and install the latest Lightpanda browser\n")
		fmt.Fprintf(os.Stderr, "  update        Update Search CLI and check Lightpanda update status\n")
		fmt.Fprintf(os.Stderr, "  uninstall     Uninstall Search CLI from current system\n")
		fmt.Fprintf(os.Stderr, "  version       Show Search CLI and Lightpanda versions\n")
		fmt.Fprintf(os.Stderr, "  --version     Show Search CLI version only\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -e <engine>   Search engine: ddg, google, brave, mojeek, hn, searx\n")
		fmt.Fprintf(os.Stderr, "  -read <url>   Read full article content from the given URL\n")
		fmt.Fprintf(os.Stderr, "  -save         Save the read article to a markdown file\n")
		fmt.Fprintf(os.Stderr, "  -panda        Force use of Lightpanda for reading\n")
		fmt.Fprintf(os.Stderr, "  -archive      Use archive.today prefix for paywalls\n")
		fmt.Fprintf(os.Stderr, "  -hn <cat>     Hacker News category: top, new, best, ask, show, job\n")
		fmt.Fprintf(os.Stderr, "  -lightpanda-only  With `update`, only check/update Lightpanda\n")
		fmt.Fprintf(os.Stderr, "  -keep-lightpanda  With `uninstall`, keep Lightpanda files\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  search \"golang news\"\n")
		fmt.Fprintf(os.Stderr, "  search -e hn -hn best\n")
		fmt.Fprintf(os.Stderr, "  search -read \"https://...\" -save\n")
		fmt.Fprintf(os.Stderr, "  search update\n")
		fmt.Fprintf(os.Stderr, "  search uninstall --keep-lightpanda\n")
	}

	engineFlag := flag.String("e", getDefaultEngine(), "Engine: ddg, google, brave, mojeek, hn, searx")
	instanceFlag := flag.String("i", "", "Searx instance URL (only for -e searx)")
	hnCatFlag := flag.String("hn", "top", "HN category: top, new, best, ask, show, job")
	readURL := flag.String("read", "", "URL to read article content from")
	archiveFlag := flag.Bool("archive", false, "Use archive.today to read the URL (for paywalls)")
	pandaFlag := flag.Bool("panda", false, "Use lightpanda headless browser for reading")
	saveFlag := flag.Bool("save", false, "Save the read article to a markdown file")
	versionFlag := flag.Bool("version", false, "Show Search CLI version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return
	}

	// Handle special commands
	if flag.NArg() > 0 {
		switch strings.ToLower(flag.Arg(0)) {
		case "setup":
			if err := util.EnsureLightpanda(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "update":
			opts, err := parseSubcommandOptions("update", flag.Args()[1:])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			if opts.LightpandaOnly {
				fmt.Println("Checking Lightpanda update status...")
				if err := util.UpdateLightpanda(); err != nil {
					fmt.Printf("Error: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("[✔] Lightpanda update check complete.")
				return
			}

			fmt.Println("Checking Search CLI latest release (skip if already up to date)...")
			if err := runInstallerAction("update", false); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Checking Lightpanda update status...")
			if err := util.UpdateLightpanda(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("[✔] Update check complete.")
			return
		case "uninstall", "unistall":
			opts, err := parseSubcommandOptions("uninstall", flag.Args()[1:])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Uninstalling Search CLI...")
			if err := runInstallerAction("uninstall", opts.KeepLightpanda); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "version":
			fmt.Printf("Search CLI: %s\n", version)
			fmt.Printf("Lightpanda: %s\n", util.GetLocalLightpandaVersion())
			return
		}
	}

	if *readURL != "" {
		// Ensure Lightpanda is installed if likely needed
		_ = util.EnsureLightpanda()

		finalURL := *readURL
		if *archiveFlag {
			finalURL = "https://archive.today/" + finalURL
		}

		var article *reader.Article
		var err error

		if *pandaFlag {
			fmt.Printf("Forcing Lightpanda for: %s...\n", finalURL)
			article, err = reader.ReadURLWithLightpanda(finalURL)
		} else {
			article, err = reader.SmartRead(finalURL)
			isPaywall := strings.Contains(finalURL, "nytimes.com") || strings.Contains(finalURL, "wsj.com") || strings.Contains(finalURL, "bloomberg.com")
			if (err != nil || len(article.Content) < 100) && isPaywall && !*archiveFlag {
				fmt.Println("\nDetected a paywalled site. Automatically retrying via archive.today...")
				archiveURL := "https://archive.today/" + finalURL
				article, err = reader.SmartRead(archiveURL)
			}
		}

		if err != nil {
			fmt.Printf("\nError: All fetch methods failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nTITLE: %s\n", article.Title)
		fmt.Println(strings.Repeat("=", 60))
		fmt.Printf("\n%s\n", article.Content)

		if *saveFlag {
			filename := sanitizeFilename(article.Title) + ".md"
			content := fmt.Sprintf("# %s\n\nSource: %s\n\n%s", article.Title, finalURL, article.Content)
			err := os.WriteFile(filename, []byte(content), 0644)
			if err != nil {
				fmt.Printf("\nError saving file: %v\n", err)
			} else {
				fmt.Printf("\n[✔] Article saved to: %s\n", filename)
			}
		}
		return
	}

	query := strings.Join(flag.Args(), " ")

	if strings.ToLower(*engineFlag) != "hn" && query == "" {
		flag.Usage()
		os.Exit(1)
	}

	var searchEngine engine.SearchEngine

	switch strings.ToLower(*engineFlag) {
	case "google":
		searchEngine = &engine.GoogleEngine{}
	case "ddg":
		searchEngine = &engine.DuckDuckGoEngine{}
	case "brave":
		searchEngine = &engine.BraveEngine{}
	case "mojeek":
		searchEngine = &engine.MojeekEngine{}
	case "hn":
		searchEngine = &engine.HackerNewsEngine{Category: *hnCatFlag}
	case "searx":
		if *instanceFlag != "" {
			searchEngine = &engine.SearxEngine{InstanceURL: *instanceFlag}
		} else {
			for _, instance := range defaultSearxInstances {
				fmt.Printf("Trying Searx instance: %s...\n", instance)
				s := &engine.SearxEngine{InstanceURL: instance}
				results, err := s.Search(query)
				if err == nil && len(results) > 0 {
					ui.PrintResults(s.Name(), query, results)
					return
				}
			}
			fmt.Println("All default Searx instances failed.")
			return
		}
	default:
		fmt.Printf("Unknown engine: %s\n", *engineFlag)
		os.Exit(1)
	}

	results, err := searchEngine.Search(query)
	if err != nil {
		primary := strings.ToLower(*engineFlag)
		if shouldFallback(primary, err) {
			for _, fallbackName := range fallbackEngineNames(primary) {
				fallbackEngine, ok := buildEngine(fallbackName, "", *hnCatFlag)
				if !ok {
					continue
				}
				fmt.Printf("Primary engine %s failed (%v). Trying fallback: %s...\n", primary, err, fallbackName)
				fallbackResults, fallbackErr := fallbackEngine.Search(query)
				if fallbackErr == nil {
					ui.PrintResults(fallbackEngine.Name(), query, fallbackResults)
					return
				}
			}
		}

		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	ui.PrintResults(searchEngine.Name(), query, results)
}

func sanitizeFilename(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	res := result.String()
	for strings.Contains(res, "--") {
		res = strings.ReplaceAll(res, "--", "-")
	}
	return strings.Trim(res, "-")
}
