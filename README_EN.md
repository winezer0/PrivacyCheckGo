# PrivacyCheckGo

High-performance sensitive information detection tool for static code analysis, fully compatible with HAE rule format.

## Project Overview

PrivacyCheckGo is a high-performance static code audit tool designed to extract sensitive information and privacy data from source code, web source code, mini-program source code, and other text files. This project is fully compatible with HAE (HaE - Highlighter and Extractor) rule format and has been comprehensively optimized for performance and functionality.

## Key Features

### 🚀 **High Performance**
- **Multi-threaded concurrent scanning**: Fully utilizes multi-core CPUs for significantly improved scanning speed
- **Memory optimization**: Supports large file chunked processing to reduce memory usage
- **Smart caching**: Supports resume scanning to avoid re-analyzing large projects
- **Cross-platform support**: Native support for Windows and Linux platforms

### 🔍 **Powerful Detection Capabilities**
- **HAE rule compatibility**: Directly reuses existing HAE rules without conversion
- **Multiple sensitive information detection**:
  - Personal information: Email, ID card numbers, mobile numbers, etc.
  - System information: Internal IPs, MAC addresses, file paths, etc.
  - Sensitive information: API keys, password fields, cloud service keys, etc.
- **Flexible rule filtering**: Supports filtering by rule group, rule name, and sensitivity level
- **Smart encoding detection**: Automatically detects file encoding (UTF-8, GBK, GB2312, etc.)

## Quick Start

### Installation

### Build Requirements
- Go 1.20+
- No CGO dependencies, supports cross-compilation

#### Option 1: Download Pre-compiled Binary
Download the binary for your platform from the [Releases page](../../releases):
- Windows x64: `privacycheck-windows-x64.exe`
- Linux x64: `privacycheck-linux-x64`

#### Option 2: Build from Source
```bash
# Clone the project
git clone https://github.com/your-repo/PrivacyCheckGo.git
cd PrivacyCheckGo

# For Windows users
build.bat

# For Linux users
./build.sh
```

### Basic Usage

```bash
# Scan a single file
./privacycheck -p /path/to/file.js

# Scan entire directory
./privacycheck -p /path/to/project

# Use custom rules file
./privacycheck -p /path/to/project -r custom_rules.yaml

# Detect only sensitive information
./privacycheck -p /path/to/project -S

# Output in CSV format
./privacycheck -p /path/to/project -f csv

# Enable caching (recommended for large projects)
./privacycheck -p /path/to/project -s
```

## Command Line Parameters

### Basic Parameters
- `-p, --project-path`: Target file or directory to scan (required)
- `-r, --rules`: Rules file path (default: config.yaml)
- `-n, --project-name`: Project name, affects output filename and cache filename

### Performance Parameters
- `-w, --workers`: Number of worker threads (default: 8)
- `--ls`: File size limit in MB (default: 5, 0 means unlimited)
- `--lc`: Chunk reading threshold in MB (default: 5, 0 means disabled)
- `-s, --save-cache`: Enable caching functionality

### Filtering Parameters
- `--ee`: List of file extensions to exclude
- `--ep`: List of path keywords to exclude
- `-S, --sensitive-only`: Detect only sensitive information
- `-N, --filter-names`: Filter by rule names
- `-G, --filter-groups`: Filter by rule groups

### Output Parameters
- `-o, --output-file`: Output file path
- `-f, --output-format`: Output format (json/csv, default: json)
- `-g, --output-group`: Output by rule groups to separate files
- `-O, --output-keys`: Specify output fields
- `-F, --format-results`: Format output results (default: enabled)
- `-b, --block-matches`: Blacklist keyword filtering

### Logging Parameters
- `--ll`: Log level (debug/info/warn/error, default: info)
- `--lf`: Log file path
- `--cf`: Console log format (default: TLM)

### Utility Parameters
- `-h, --help`: Show help information

## Configuration File Format

PrivacyCheckGo is fully compatible with HAE rule format. Configuration file example:

```yaml
rules:
  - group: Sensitive Information
    rule:
      - name: Cloud Key
        loaded: true
        f_regex: (((access)(|-|_)(key)(|-|_)(id|secret))|(LTAI[a-z0-9]{12,20}))
        sensitive: true
        context_left: 50
        context_right: 50
```

### Rule Parameters
- `name`: Rule name (required)
- `f_regex`: Regular expression pattern (required)
- `sensitive`: Whether it's sensitive information (default: false)
- `loaded`: Whether to enable the rule (default: true)
- `context_left`: Number of context characters to expand left
- `context_right`: Number of context characters to expand right

### Important Notes
- **Case-insensitive matching**: All regex patterns are matched case-insensitively by default
- **Context extraction**: When context_left/right > 0, the tool extracts surrounding context for better analysis
- **Rule grouping**: Rules are organized by groups for better management and output organization

## Usage Examples

### Basic Scanning
```bash
# Scan current directory
./privacycheck -p .

# Scan specific project with caching enabled
./privacycheck -p /path/to/large-project -s -n my-project
```

## Disclaimer

This tool is intended for legitimate security testing and code auditing purposes only. Users should comply with relevant laws and regulations and must not use this tool for illegal purposes. The developers assume no responsibility for any consequences arising from misuse of this tool.
