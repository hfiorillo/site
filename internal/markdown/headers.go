package markdown

import (
	"bufio"
	"strings"

	"github.com/hfiorillo/site/models"
)

// TODO: Maybe only have the mid and bottom level headers
// ParseHeaders extracts headers from markdown text and returns an assembled Headers struct
func parseHeaders(content []byte) models.Headers {

	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	var result models.Headers

	var currentTop *struct {
		Headers  string
		MidLevel []struct {
			Headers     string
			BottomLevel []struct {
				Headers string
			}
		}
	}
	var currentMid *struct {
		Headers     string
		BottomLevel []struct {
			Headers string
		}
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "# ") {
			// Top-level header
			top := struct {
				Headers  string
				MidLevel []struct {
					Headers     string
					BottomLevel []struct {
						Headers string
					}
				}
			}{
				Headers: strings.TrimPrefix(line, "# "),
			}
			result.TopLevel = append(result.TopLevel, top)
			currentTop = &result.TopLevel[len(result.TopLevel)-1]
			currentMid = nil // Reset mid-level pointer
		} else if strings.HasPrefix(line, "## ") && currentTop != nil {
			// Mid-level header
			mid := struct {
				Headers     string
				BottomLevel []struct {
					Headers string
				}
			}{
				Headers: strings.TrimPrefix(line, "## "),
			}
			currentTop.MidLevel = append(currentTop.MidLevel, mid)
			currentMid = &currentTop.MidLevel[len(currentTop.MidLevel)-1]
		} else if strings.HasPrefix(line, "### ") && currentMid != nil {
			// Bottom-level header
			bottom := struct {
				Headers string
			}{
				Headers: strings.TrimPrefix(line, "### "),
			}
			currentMid.BottomLevel = append(currentMid.BottomLevel, bottom)
		}
	}

	return result
}
