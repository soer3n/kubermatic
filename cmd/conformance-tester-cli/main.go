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
	"encoding/json"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"k8c.io/kubermatic/v2/cmd/conformance-tester-cli/internal/form"
	"k8c.io/kubermatic/v2/cmd/conformance-tester-cli/internal/runner"
)

func main() {
	// Create form model
	formModel := form.NewFormModel()

	// Run the form using bubbletea
	p := tea.NewProgram(formModel)
	_, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Post-process the form data
	if err := formModel.FormData.PostProcess(); err != nil {
		log.Fatal(err)
	}

	// Generate output configuration
	out := formModel.FormData.Config.ToOutputConfig(*formModel.FormData.Secrets)

	// If the user opted to run the tests, execute them now.
	if formModel.FormData.RunTests {
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
