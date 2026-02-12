/*
                  Kubermatic Enterprise Read-Only License
                         Version 1.0 ("KERO-1.0")
                     Copyright © 2025 Kubermatic GmbH

   1.	You may only view, read and display for studying purposes the source
      code of the software licensed under this license, and, to the extent
      explicitly provided under this license, the binary code.
   2.	Any use of the software which exceeds the foregoing right, including,
      without limitation, its execution, compilation, copying, modification
      and distribution, is expressly prohibited.
   3.	THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND,
      EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
      MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
      IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
      CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
      TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
      SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

   END OF TERMS AND CONDITIONS
*/

package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderCommonLayout renders content with banner, vertical centering, and help bar at bottom
func (m Model) renderCommonLayout(content string, helpWithBorder string) string {
	uiWidth := m.getUIWidth()

	// Combine banner and content, then center the entire layout
	bannerContent := styleBanner.Width(uiWidth).Render(bannerText())
	finalContent := lipgloss.JoinVertical(lipgloss.Center, bannerContent, content)

	// Vertically center the content in available space and place help bar at bottom
	availableHeight := m.terminalHeight - lipgloss.Height(helpWithBorder)
	centeredContent := lipgloss.Place(uiWidth+8, availableHeight, lipgloss.Center, lipgloss.Center, finalContent)

	contentLines := lipgloss.Height(centeredContent)
	helpBarLines := lipgloss.Height(helpWithBorder)
	totalUsedLines := contentLines + helpBarLines

	if m.terminalHeight > 0 && totalUsedLines < m.terminalHeight {
		// Add spacing to push help bar to bottom
		spacing := m.terminalHeight - totalUsedLines - 1
		if spacing > 0 {
			spacer := strings.Repeat("\n", spacing)
			return centeredContent + spacer + helpWithBorder
		}
	}

	return centeredContent + "\n" + helpWithBorder
}

// renderCommonHelpBar renders the help bar with consistent styling
func (m Model) renderCommonHelpBar(stage int) string {
	uiInnerWidth := m.getUIInnerWidth()

	helpText := helpBar(stage)
	helpContent := styleHelpBar.Width(m.terminalWidth).Render(helpText)
	helpWithBorder := styleHelpBarBorder.Width(uiInnerWidth).Render("") + "\n" + helpContent

	return helpWithBorder
}

// padContentToHeight pads content to a specific height for consistent layout
func padContentToHeight(content string, targetHeight int) []string {
	lines := strings.Split(content, "\n")
	currentHeight := len(lines)

	if currentHeight >= targetHeight {
		return lines
	}

	// Add empty lines to reach target height
	padding := make([]string, targetHeight-currentHeight)
	for i := range padding {
		padding[i] = ""
	}

	return append(lines, padding...)
}
