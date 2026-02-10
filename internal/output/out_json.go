package output

import (
	"fmt"
	"github.com/winezer0/xutils/utils"
	"privacycheck/internal/scanner"
)

// writeJSON 写入JSON文件
func (p *Output) writeJSON(filename string, results []scanner.ScanResult) error {
	if err := utils.SaveJSON(filename, results); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}
	return nil
}
