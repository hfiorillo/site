package markdown

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/hfiorillo/site/models"
)

func ParseHeaders(content []byte) models.Headers {
	scanner := bufio.NewScanner(bytes.NewReader(content))

	var result models.Headers

	var currentTop *models.TopLevelHeader
	var currentMid *models.MidLevelHeader

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "# ") {
			headerText := cleanHeader(strings.TrimPrefix(line, "# "))
			top := models.TopLevelHeader{Headers: headerText}
			result.TopLevel = append(result.TopLevel, top)
			currentTop = &result.TopLevel[len(result.TopLevel)-1]
			currentMid = nil
		} else if strings.HasPrefix(line, "## ") {
			if currentTop == nil {
				continue
			}
			headerText := cleanHeader(strings.TrimPrefix(line, "## "))
			mid := models.MidLevelHeader{Headers: headerText}
			currentTop.MidLevel = append(currentTop.MidLevel, mid)
			currentMid = &currentTop.MidLevel[len(currentTop.MidLevel)-1]
		} else if strings.HasPrefix(line, "### ") {
			if currentMid == nil {
				continue
			}
			headerText := cleanHeader(strings.TrimPrefix(line, "### "))
			bottom := models.BottomLevelHeader{Headers: headerText}
			currentMid.BottomLevel = append(currentMid.BottomLevel, bottom)
		}
	}

	return result
}

func cleanHeader(text string) string {
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "__", "")
	text = strings.ReplaceAll(text, "`", "")
	return strings.TrimSpace(text)
}
