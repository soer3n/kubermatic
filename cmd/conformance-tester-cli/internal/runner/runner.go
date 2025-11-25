package runner

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"

	"gopkg.in/yaml.v3"
	"k8c.io/kubermatic/v2/cmd/conformance-tester-cli/internal/config"
)

// Run executes the Ginkgo test suite for each selected provider in parallel.
func Run(cfg config.OutputConfig) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(cfg.Providers))

	for providerName, settings := range cfg.Providers {
		wg.Add(1)
		go func(providerName string, settings config.ProviderSettings) {
			defer wg.Done()

			// Create a Ginkgo-compatible config for this provider.
			ginkgoCfg := config.GinkgoOutputConfig{
				NamePrefix:          cfg.NamePrefix,
				Providers:           []string{providerName},
				Releases:            cfg.Releases,
				EnableDistributions: cfg.EnableDistributions,
				TestSettings:        settings.TestSettings,
				DeleteCluster:       cfg.DeleteClusterAfterTests,
				NodeCount:           cfg.NodeCount,
				ReportsRoot:         "_reports",
				LogDirectory:        "_logs",
				KubermaticSeedName:  cfg.KubermaticSeedName,
				Secrets:             cfg.Secrets,
			}

			// Create a temporary YAML file for the Ginkgo runner.
			tmpfile, err := os.CreateTemp("", "ginkgo-config-*.yaml")
			if err != nil {
				errChan <- fmt.Errorf("failed to create temp config for %s: %w", providerName, err)
				return
			}
			defer os.Remove(tmpfile.Name())

			encoder := yaml.NewEncoder(tmpfile)
			if err := encoder.Encode(ginkgoCfg); err != nil {
				errChan <- fmt.Errorf("failed to write ginkgo config for %s: %w", providerName, err)
				return
			}
			if err := tmpfile.Close(); err != nil {
				errChan <- fmt.Errorf("failed to close temp config for %s: %w", providerName, err)
				return
			}

			log.Printf("Starting tests for provider: %s", providerName)

			// Execute the tests using `go test`.
			cmd := exec.Command("go", "test", "-v", "./cmd/conformance-tester/pkg/ginkgo/...")
			cmd.Env = append(os.Environ(), "CONFORMANCE_TESTER_CONFIG_FILE="+tmpfile.Name())

			// Stream output to the console.
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				errChan <- fmt.Errorf("failed to get stdout pipe for %s: %w", providerName, err)
				return
			}
			stderr, err := cmd.StderrPipe()
			if err != nil {
				errChan <- fmt.Errorf("failed to get stderr pipe for %s: %w", providerName, err)
				return
			}

			if err := cmd.Start(); err != nil {
				errChan <- fmt.Errorf("failed to start tests for %s: %w", providerName, err)
				return
			}

			// Use a separate goroutine to stream output to avoid blocking.
			var streamWg sync.WaitGroup
			streamWg.Add(2)
			go streamOutput(stdout, &streamWg, providerName)
			go streamOutput(stderr, &streamWg, providerName)
			streamWg.Wait()

			if err := cmd.Wait(); err != nil {
				errChan <- fmt.Errorf("tests failed for %s: %w", providerName, err)
				return
			}

			log.Printf("Successfully finished tests for provider: %s", providerName)
		}(providerName, settings)
	}

	wg.Wait()
	close(errChan)

	// Check for any errors that occurred during the test runs.
	var finalErr error
	for err := range errChan {
		if finalErr == nil {
			finalErr = err
		} else {
			finalErr = fmt.Errorf("%v; %w", finalErr, err)
		}
	}

	return finalErr
}

// streamOutput reads from an io.Reader and prints lines prefixed with the provider name.
func streamOutput(r io.Reader, wg *sync.WaitGroup, prefix string) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		log.Printf("[%s] %s", prefix, scanner.Text())
	}
}
