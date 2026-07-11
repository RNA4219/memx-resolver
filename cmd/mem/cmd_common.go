package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/RNA4219/memx-resolver/v2/api"
	"github.com/RNA4219/memx-resolver/v2/db"
	"github.com/RNA4219/memx-resolver/v2/service"
)

// -------------------- common flags --------------------

type commonFlags struct {
	apiURL string
	json   bool

	short     string
	journal   string
	knowledge string
	archive   string
	resolver  string
}

func (c *commonFlags) bind(fs *flag.FlagSet) {
	fs.StringVar(&c.apiURL, "api-url", "", "HTTP API base URL (if set, CLI uses HTTP client)")
	fs.BoolVar(&c.json, "json", false, "output JSON")
	fs.StringVar(&c.short, "short", "short.db", "path to short.db")
	fs.StringVar(&c.journal, "journal", "journal.db", "path to journal.db")
	fs.StringVar(&c.knowledge, "knowledge", "knowledge.db", "path to knowledge.db")
	fs.StringVar(&c.archive, "archive", "archive.db", "path to archive.db")
	fs.StringVar(&c.resolver, "resolver", "", "path to resolver.db (optional, defaults to short.db)")
}

func (c *commonFlags) paths() db.Paths {
	return db.Paths{
		Short:     c.short,
		Journal:   c.journal,
		Knowledge: c.knowledge,
		Archive:   c.archive,
		Resolver:  c.resolver,
	}
}

func makeClient(ctx context.Context, cf *commonFlags) (api.Client, func(), error) {
	if cf.apiURL != "" {
		return api.NewHTTPClient(cf.apiURL), func() {}, nil
	}
	svc, err := service.New(cf.paths())
	if err != nil {
		return nil, nil, err
	}
	if err := attachOpenAIFromEnv(svc); err != nil {
		_ = svc.Close()
		return nil, nil, err
	}
	cleanup := func() { _ = svc.Close() }
	return api.NewInProcClient(svc), cleanup, nil
}

func attachOpenAIFromEnv(svc *service.Service) error {
	return svc.ConfigureLLMsFromEnv()
}

func printJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// -------------------- env loader --------------------

func loadDotEnvFromHierarchy() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	path, ok := findDotEnvPath(cwd)
	if !ok {
		return nil
	}
	return loadDotEnvFile(path)
}

func findDotEnvPath(start string) (string, bool) {
	dir := start
	for {
		candidate := filepath.Join(dir, ".env")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func loadDotEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := parseDotEnvValue(parts[1])
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func parseDotEnvValue(raw string) string {
	value := strings.TrimSpace(raw)
	if len(value) >= 2 {
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}
	}
	return value
}

// -------------------- helper types --------------------

type multiStringFlag []string

func (m *multiStringFlag) String() string { return strings.Join(*m, ",") }
func (m *multiStringFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

type flagsStringSlice []string

func (f *flagsStringSlice) String() string { return strings.Join(*f, ",") }
func (f *flagsStringSlice) Set(v string) error {
	*f = append(*f, v)
	return nil
}

// -------------------- helper functions --------------------

func readAllStdin() string {
	b := &strings.Builder{}
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		b.WriteString(s.Text())
		b.WriteByte('\n')
	}
	return b.String()
}

func parseStores(s string) []string {
	if s == "" {
		return []string{"short"}
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// -------------------- setup --------------------

func runSetup() {
	log.SetFlags(0)
	if err := loadDotEnvFromHierarchy(); err != nil {
		log.Printf("warning: failed to load .env: %v", err)
	}
}
