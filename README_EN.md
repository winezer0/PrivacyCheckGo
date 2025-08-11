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
- **Direct file path processing**: Optimized scanner that processes files directly without pre-conversion overhead

### 🔍 **Powerful Detection Capabilities**
- **HAE rule compatibility**: Directly reuses existing HAE rules without conversion
- **Multiple sensitive information detection**:
  - Personal information: Email, ID card numbers, mobile numbers, etc.
  - System information: Internal IPs, MAC addresses, file paths, etc.
  - Sensitive information: API keys, password fields, cloud service keys, etc.
- **Flexible rule filtering**: Supports filtering by rule group, rule name, and sensitivity level
- **Smart encoding detection**: Automatically detects file encoding (UTF-8, GBK, GB2312, etc.)
- **Case-insensitive matching**: Default case-insensitive regex matching for better detection

### 📊 **Rich Output Options**
- **Multiple output formats**: Supports JSON and CSV formats
- **Flexible result grouping**: Can output by rule groups to separate files
- **Custom output fields**: Output only the fields you need
- **Result filtering**: Supports blacklist keyword filtering
- **Format options**: Automatically cleans special characters in results

### ⚙️ **Easy to Use**
- **Smart configuration management**: Automatically generates default configuration when config file doesn't exist
- **Detailed logging system**: Supports multi-level logging and file output
- **Real-time progress display**: Shows real-time progress and estimated remaining time during scanning
- **Command-line friendly**: Rich command-line parameters with intelligent abbreviations

## Quick Start

### Installation

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
  - group: People Information
    rule:
      - name: Email
        loaded: true
        f_regex: (([a-z0-9]+[_|\.])*[a-z0-9]+@([a-z0-9]+[-|_|\.])*[a-z0-9]+\.[a-z]{2,5})
        sensitive: false
        context_left: 0
        context_right: 0
      - name: Chinese IDCard
        loaded: true
        f_regex: '[^0-9]((\d{8}(0\d|10|11|12)([0-2]\d|30|31)\d{3}$)|(\d{6}(18|19|20)\d{2}(0[1-9]|10|11|12)([0-2]\d|30|31)\d{3}(\d|X|x)))[^0-9]'
        sensitive: true
        context_left: 50
        context_right: 50
  - group: System Information
    rule:
      - name: Internal IP Address
        loaded: true
        f_regex: '[^0-9]((127\.0\.0\.1)|(10\.\d{1,3}\.\d{1,3}\.\d{1,3})|(172\.((1[6-9])|(2\d)|(3[01]))\.\d{1,3}\.\d{1,3})|(192\.168\.\d{1,3}\.\d{1,3}))'
        sensitive: true
        context_left: 50
        context_right: 50
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

### Advanced Filtering
```bash
# Detect only sensitive information
./privacycheck -p /path/to/project -S

# Detect only rules containing "password"
./privacycheck -p /path/to/project -N password

# Detect only rules from "Sensitive Information" group
./privacycheck -p /path/to/project -G "Sensitive Information"

# Exclude specific file types
./privacycheck -p /path/to/project --ee .log,.tmp,.bak

# Exclude specific paths
./privacycheck -p /path/to/project --ep /tmp,/cache,node_modules
```

### Output Customization
```bash
# Output in CSV format
./privacycheck -p /path/to/project -f csv

# Output by rule groups separately
./privacycheck -p /path/to/project -g

# Output only specified fields
./privacycheck -p /path/to/project -O file,rule_name,match

# Filter results containing specific keywords
./privacycheck -p /path/to/project -b "test","example"
```

### Performance Optimization
```bash
# Use 8 threads for scanning
./privacycheck -p /path/to/project -w 8

# Limit scan file size to 10MB
./privacycheck -p /path/to/project --ls 10

# Set chunk reading threshold to 20MB (large file optimization)
./privacycheck -p /path/to/project --lc 20

# Disable chunk reading (high performance mode)
./privacycheck -p /path/to/project --lc 0

# Memory optimization mode (1MB chunk threshold)
./privacycheck -p /path/to/project --lc 1
```

## Chunk Reading Feature

### Smart Memory Management
PrivacyCheckGo supports intelligent chunk reading functionality to effectively control memory usage:

#### How It Works
- **Automatic switching**: Automatically selects reading strategy based on file size
- **Chunk processing**: Large files are processed in 1MB chunks to maintain low memory usage
- **Line integrity**: Ensures no truncation in the middle of lines, maintaining matching accuracy
- **Position accuracy**: Correctly calculates file position and line number for each match result

#### Configuration Options
```bash
# Default configuration (5MB threshold)
./privacycheck -p /path/to/project

# Custom threshold (10MB)
./privacycheck -p /path/to/project --lc 10

# Memory optimization (1MB threshold)
./privacycheck -p /path/to/project --lc 1

# Disable chunk reading (full file reading)
./privacycheck -p /path/to/project --lc 0
```

#### Usage Recommendations
- **Memory-constrained environments**: Set smaller threshold (1-2MB)
- **High-performance environments**: Set larger threshold (10-20MB) or disable chunk reading
- **Default configuration**: 5MB threshold suits most scenarios
- **Large file projects**: Enable chunk reading to significantly reduce memory usage

## Technical Architecture

### Project Structure
```
PrivacyCheckGo/
├── internal/
│   ├── baserule/  # Rule system
│   ├── scanner/   # Scanning engine
│   └── output/    # Output processing
├── pkg/           # Utility packages
│   ├── fileutils/ # File processing utilities
│   └── logging/   # Logging system
├── main.go        # Main program entry
├── config.yaml    # Default rule configuration
├── build.bat      # Windows build script
├── build.sh       # Linux build script
└── README.md      # Project documentation
```

### Core Components
1. **Rule System (baserule)**: HAE-compatible rule loading, validation, and filtering
2. **Scanning Engine (scanner)**: Multi-threaded scan scheduling, rule matching, and result collection
   - **Optimized file processing**: Direct filepath processing without pre-conversion overhead
   - **Lazy encoding detection**: Encoding detection only when needed during actual file reading
   - **Smart caching**: Intelligent cache management with resume scanning support
3. **File Processing (fileutils)**: File discovery, encoding detection, and chunk reading
4. **Output Processing (output)**: Result formatting, filtering, and multi-format output
5. **Logging System (logging)**: Structured logging with multi-level output

### Performance Optimizations
- **Direct filepath processing**: Scanner processes files directly without FileInfo pre-conversion
- **Lazy evaluation**: File size and encoding detection only when actually needed
- **Memory-efficient chunking**: Large files processed in 1MB chunks with accurate line tracking
- **Case-insensitive regex**: Default case-insensitive matching for better detection coverage

## Performance Comparison

Compared to the Python version, the Go version has significant improvements in the following areas:

| Feature | Python Version | Go Version | Improvement |
|---------|----------------|------------|-------------|
| Startup Speed | ~2 seconds | ~0.1 seconds | 20x |
| Memory Usage | High | Low | 50%+ |
| Scanning Speed | Baseline | 3-5x | 3-5x |
| Concurrency | GIL Limited | True Concurrency | Significant |
| Deployment | Requires Python Environment | Single Binary | Massive Improvement |
| File Processing | Pre-conversion Overhead | Direct Processing | Faster Startup |
| Regex Matching | Case-sensitive by default | Case-insensitive by default | Better Detection |

## Development Guide

### Dependencies
- `github.com/jessevdk/go-flags`: Command-line argument parsing
- `go.uber.org/zap`: High-performance logging library
- `golang.org/x/text`: Text encoding processing
- `gopkg.in/yaml.v3`: YAML configuration parsing

### Build Requirements
- Go 1.20+
- No CGO dependencies, supports cross-compilation

### Build Instructions
The project provides automated build scripts:
- Windows: Run `build.bat`
- Linux: Run `./build.sh`

The build scripts automatically:
1. Check Go environment
2. Download dependencies
3. Cross-compile for Windows and Linux
4. Generate optimized binaries

### Recent Optimizations
- **Direct filepath processing**: Removed FileInfo pre-conversion overhead
- **Lazy encoding detection**: Encoding detection only when files are actually read
- **Case-insensitive regex**: Default case-insensitive matching for better detection
- **Memory optimization**: Reduced memory footprint through direct file processing

## Contributing

Welcome to submit Issues and Pull Requests!

### Development Environment Setup
```bash
git clone https://github.com/your-repo/PrivacyCheckGo.git
cd PrivacyCheckGo
go mod tidy
```

### Code Standards
- Use `gofmt` to format code
- Follow Go language best practices
- Add necessary comments and documentation
- Write unit tests

## Disclaimer

This tool is intended for legitimate security testing and code auditing purposes only. Users should comply with relevant laws and regulations and must not use this tool for illegal purposes. The developers assume no responsibility for any consequences arising from misuse of this tool.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact

If you have questions or suggestions, please contact us through:
- Submit Issues: [GitHub Issues](../../issues)
- Email: your-email@example.com

---

**PrivacyCheckGo - Making Code Security Detection Faster and Stronger!**
