package output

import (
	"fmt"
	"privacycheck/internal/scanner"
	"privacycheck/pkg/fileutils"
	"reflect"
)

// writeCSV 写入CSV文件
func (p *Output) writeCSV(filename string, results []scanner.ScanResult) error {
	if len(results) == 0 {
		return nil
	}

	// 准备表头
	headers := p.getCSVHeaders(results[0])

	// 准备数据记录
	var records [][]string
	for _, result := range results {
		record := p.resultToCSVRecord(result, headers)
		records = append(records, record)
	}

	// 使用fileutils写入CSV
	if err := fileutils.WriteCSVFile(filename, headers, records); err != nil {
		return fmt.Errorf("failed to write CSV file: %w", err)
	}

	return nil
}

// getCSVHeaders 获取CSV表头
func (p *Output) getCSVHeaders(result scanner.ScanResult) []string {
	// 如果指定了输出字段，使用指定的字段顺序
	if len(p.OutputKeys) > 0 {
		return p.OutputKeys
	}

	// 否则使用默认字段顺序
	return p.getDefaultHeaders(result)
}

// getDefaultHeaders 获取默认表头（通过反射）
func (p *Output) getDefaultHeaders(result scanner.ScanResult) []string {
	var headers []string
	v := reflect.ValueOf(result)
	t := reflect.TypeOf(result)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			headers = append(headers, jsonTag)
		}
	}

	return headers
}

// resultToCSVRecord 将结果转换为CSV记录
func (p *Output) resultToCSVRecord(result scanner.ScanResult, headers []string) []string {
	record := make([]string, len(headers))

	for i, header := range headers {
		record[i] = p.getFieldValue(result, header)
	}

	return record
}

// getFieldValue 获取字段值
func (p *Output) getFieldValue(result scanner.ScanResult, fieldName string) string {
	switch fieldName {
	case "file":
		return result.File
	case "group":
		return result.Group
	case "rule_name":
		return result.RuleName
	case "match":
		return result.Match
	case "context":
		return result.Context
	case "position":
		return fmt.Sprintf("%d", result.Position)
	case "line_number":
		return fmt.Sprintf("%d", result.LineNumber)
	case "sensitive":
		return fmt.Sprintf("%t", result.Sensitive)
	default:
		return ""
	}
}
