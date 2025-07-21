package output

import (
	"fmt"
	"privacycheck/pkg/fileutils"
	"privacycheck/scanner"
)

// writeJSON 写入JSON文件
func (p *Output) writeJSON(filename string, results []scanner.ScanResult) error {
	if err := fileutils.WriteJSONFile(filename, results); err != nil {
		return fmt.Errorf("写入JSON文件失败: %w", err)
	}
	return nil
}
