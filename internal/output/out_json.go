package output

import (
	"fmt"
	"privacycheck/internal/scanner"
	"privacycheck/pkg/fileutils"
)

// writeJSON 写入JSON文件
func (p *Output) writeJSON(filename string, results []scanner.ScanResult) error {
	if err := fileutils.WriteJSONFile(filename, results); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}
	return nil
}
