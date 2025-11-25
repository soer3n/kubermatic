/*
Copyright 2025 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"k8c.io/kubermatic/v2/cmd/conformance-tester-cli/internal/form"
	"k8c.io/kubermatic/v2/cmd/conformance-tester-cli/internal/runner"
)

func displaySplashScreen() {
	// Clear screen (works on Unix-like systems)
	fmt.Print("\033[H\033[2J")
	fmt.Println()
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                                ║")
	fmt.Println("║                    CONFORMANCE TESTER                          ║")
	fmt.Println("║                                                                ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println()
	fmt.Println("Press Enter to continue...")
	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Display splash screen
	displaySplashScreen()

	// Create form data
	formData := form.NewFormData()

	// Build and run the form
	formUI := formData.BuildForm()
	if err := formUI.Run(); err != nil {
		log.Fatal(err)
	}

	// Post-process the form data
	if err := formData.PostProcess(); err != nil {
		log.Fatal(err)
	}

	// Generate output configuration
	out := formData.Config.ToOutputConfig(*formData.Secrets)

	// If the user opted to run the tests, execute them now.
	if formData.RunTests {
		if err := runner.Run(out); err != nil {
			log.Fatalf("Test execution failed: %v", err)
		}
		fmt.Println("All tests completed successfully!")
		return
	}

	// Otherwise, just print the configuration as JSON.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	fmt.Println()
	fmt.Println("Configuration:")
	if err := enc.Encode(out); err != nil {
		log.Fatal(err)
	}
}
