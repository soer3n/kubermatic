//go:build ee

/*
                  Kubermatic Enterprise Read-Only License
                         Version 1.0 ("KERO-1.0”)
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

package common

import "github.com/kyverno/kyverno/api/kyverno"

func KyvernoLabels(component string) map[string]string {
	return map[string]string{
		kyverno.LabelAppComponent:    component,
		"app.kubernetes.io/instance": kyverno.ValueKyvernoApp,
		"app.kubernetes.io/part-of":  kyverno.ValueKyvernoApp,
		"app.kubernetes.io/version":  KyvernoVersion,
	}
}

func KyvernoSelectorLabels(component string) map[string]string {
	return map[string]string{
		kyverno.LabelAppComponent:    component,
		"app.kubernetes.io/instance": kyverno.ValueKyvernoApp,
		"app.kubernetes.io/part-of":  kyverno.ValueKyvernoApp,
	}
}
