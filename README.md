# PrivacyCheck Go版本

基于Python版本PrivacyCheck重新实现的Go版本，兼容HAE规则格式的静态代码敏感信息检测工具。

## 项目简介

PrivacyCheck Go版本是一个高性能的静态代码审计工具，专门用于从源代码、网页源码、小程序源码等文本文件中提取敏感信息和隐私数据。本项目完全兼容HAE（HaE - Highlighter and Extractor）规则格式，并在性能和功能上进行了全面优化。

## 主要特性

### 🚀 **高性能**
- **多线程并发扫描**：充分利用多核CPU，显著提升扫描速度
- **内存优化**：支持大文件分块处理，降低内存占用
- **智能缓存**：支持断点续扫，避免重复分析大型项目
- **跨平台支持**：原生支持Windows和Linux平台

### 🔍 **强大的检测能力**
- **兼容HAE规则**：直接复用现有HAE规则，无需转换
- **多种敏感信息检测**：
  - 个人信息：邮箱、身份证号、手机号等
  - 系统信息：内网IP、MAC地址、文件路径等
  - 敏感信息：API密钥、密码字段、云服务密钥等
- **灵活的规则过滤**：支持按规则组、规则名称、敏感级别过滤
- **智能编码检测**：自动识别文件编码（UTF-8、GBK、GB2312等）

### 📊 **丰富的输出选项**
- **多种输出格式**：支持JSON和CSV格式
- **灵活的结果分组**：可按规则组分别输出
- **自定义输出字段**：仅输出需要的字段
- **结果过滤**：支持黑名单关键字过滤
- **格式化选项**：自动清理结果中的特殊字符

### ⚙️ **易于使用**
- **智能配置管理**：配置文件不存在时自动生成默认配置
- **详细的日志系统**：支持多级别日志和文件输出
- **实时进度显示**：扫描过程中实时显示进度和预计剩余时间
- **命令行友好**：丰富的命令行参数和智能缩写

## 快速开始

### 安装

#### 方式一：下载预编译二进制文件
从[Releases页面](../../releases)下载对应平台的二进制文件：
- Windows x64: `privacycheck-windows-x64.exe`
- Linux x64: `privacycheck-linux-x64`

#### 方式二：从源码编译
```bash
# 克隆项目
git clone https://github.com/your-repo/PrivacyCheckGo.git
cd PrivacyCheckGo

# Windows用户
build.bat

# Linux用户
./build.sh
```

### 基本使用

```bash
# 扫描单个文件
./privacycheck -p /path/to/file.js

# 扫描整个目录
./privacycheck -p /path/to/project

# 使用自定义规则文件
./privacycheck -p /path/to/project -r custom_rules.yaml

# 仅检测敏感信息
./privacycheck -p /path/to/project -S

# 输出为CSV格式
./privacycheck -p /path/to/project -f csv

# 启用缓存（推荐大项目使用）
./privacycheck -p /path/to/project -s
```

## 命令行参数

### 基础参数
- `-p, --project-path`: 待扫描的目标文件或目录（必需）
- `-r, --rules`: 规则文件路径（默认：config.yaml）
- `-n, --project-name`: 项目名称，影响输出文件名和缓存文件名

### 性能参数
- `-w, --workers`: 工作线程数量（默认：CPU核心数）
- `--ls`: 文件大小限制（MB，默认：5，0表示无限制）
- `--cl`: 分块读取阈值（MB，默认：5，0表示禁用分块读取）
- `-s, --save-cache`: 启用缓存功能

### 过滤参数
- `--ee`: 排除的文件扩展名列表
- `--ep`: 排除的路径关键字列表
- `-S, --sensitive-only`: 仅检测敏感信息
- `-N, --filter-names`: 按规则名称过滤
- `-G, --filter-groups`: 按规则组过滤

### 输出参数
- `-o, --output-file`: 输出文件路径
- `-f, --output-format`: 输出格式（json/csv，默认：json）
- `-g, --output-group`: 按规则组分别输出
- `-O, --output-keys`: 指定输出字段
- `-F, --format-results`: 格式化输出结果（默认：true）
- `-b, --block-matches`: 黑名单关键字过滤

### 日志参数
- `--log-level`: 日志级别（debug/info/warn/error，默认：info）
- `--log-file`: 日志文件路径
- `--log-format`: 控制台日志格式（默认：TLM）

### 工具参数
- `-h, --help`: 显示帮助信息

## 配置文件格式

PrivacyCheck Go版本完全兼容HAE规则格式。配置文件示例：

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
      - name: API Key
        loaded: true
        f_regex: '((api|key|token|secret|auth)[\w\-_]*[\s]*[:=][\s]*[''"]?[a-zA-Z0-9\-_]{16,}[''"]?)'
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
```

### 规则参数说明
- `name`: 规则名称（必需）
- `f_regex`: 正则表达式（必需）
- `sensitive`: 是否为敏感信息（默认：false）
- `loaded`: 是否启用规则（默认：true）
- `context_left`: 向左扩展的上下文字符数
- `context_right`: 向右扩展的上下文字符数

## 使用示例

### 基础扫描
```bash
# 扫描当前目录
./privacycheck -p .

# 扫描指定项目，启用缓存
./privacycheck -p /path/to/large-project -s -n my-project
```

### 高级过滤
```bash
# 仅检测敏感信息
./privacycheck -p /path/to/project -S

# 仅检测包含"password"的规则
./privacycheck -p /path/to/project -N password

# 仅检测"Sensitive Information"组的规则
./privacycheck -p /path/to/project -G "Sensitive Information"

# 排除特定文件类型
./privacycheck -p /path/to/project --ee .log,.tmp,.bak

# 排除特定路径
./privacycheck -p /path/to/project --ep /tmp,/cache,node_modules
```

### 输出定制
```bash
# 输出为CSV格式
./privacycheck -p /path/to/project -f csv

# 按规则组分别输出
./privacycheck -p /path/to/project -g

# 仅输出指定字段
./privacycheck -p /path/to/project -O file,rule_name,match

# 过滤包含特定关键字的结果
./privacycheck -p /path/to/project -b "test","example"
```

### 性能优化
```bash
# 使用8个线程扫描
./privacycheck -p /path/to/project -w 8

# 限制扫描文件大小为10MB
./privacycheck -p /path/to/project --ls 10

# 设置分块读取阈值为20MB（大文件优化）
./privacycheck -p /path/to/project --cl 20

# 禁用分块读取（高性能模式）
./privacycheck -p /path/to/project --cl 0

# 内存优化模式（1MB分块阈值）
./privacycheck -p /path/to/project --cl 1
```

## 分块读取功能

### 智能内存管理
PrivacyCheck Go版本支持智能分块读取功能，可以有效控制内存使用：

#### 工作原理
- **自动切换**：根据文件大小自动选择读取策略
- **分块处理**：大文件按1MB块进行处理，保持低内存占用
- **行完整性**：确保不会在行中间截断，保持匹配准确性
- **位置精确**：正确计算每个匹配结果的文件位置和行号

#### 配置选项
```bash
# 默认配置（5MB阈值）
./privacycheck -p /path/to/project

# 自定义阈值（10MB）
./privacycheck -p /path/to/project --cl 10

# 内存优化（1MB阈值）
./privacycheck -p /path/to/project --cl 1

# 禁用分块读取（全量读取）
./privacycheck -p /path/to/project --cl 0
```

#### 使用建议
- **内存受限环境**：设置较小阈值（1-2MB）
- **高性能环境**：设置较大阈值（10-20MB）或禁用分块读取
- **默认配置**：5MB阈值适合大多数场景
- **大文件项目**：启用分块读取可显著降低内存占用

## 技术架构

### 项目结构
```
PrivacyCheckGo/
├── baserule/      # 规则系统
├── scanner/       # 扫描引擎
├── output/        # 输出处理
├── pkg/           # 工具包
│   ├── fileutils/ # 文件处理工具
│   └── logging/   # 日志系统
├── internal/      # 内部模块
│   └── embeds/    # 嵌入资源
├── main.go        # 主程序入口
├── config.yaml    # 默认规则配置
├── build.bat      # Windows构建脚本
├── build.sh       # Linux构建脚本
└── README.md      # 项目文档
```

### 核心组件
1. **规则系统（baserule）**：HAE兼容的规则加载、验证和过滤
2. **扫描引擎（scanner）**：多线程扫描调度、规则匹配和结果收集
3. **文件处理（fileutils）**：文件发现、编码检测、分块读取
4. **输出处理（output）**：结果格式化、过滤和多格式输出
5. **缓存系统**：智能缓存管理，支持断点续扫
6. **日志系统（logging）**：结构化日志记录和多级别输出

## 性能对比

与Python版本相比，Go版本在以下方面有显著提升：

| 特性 | Python版本 | Go版本 | 提升 |
|------|------------|--------|------|
| 启动速度 | ~2秒 | ~0.1秒 | 20x |
| 内存占用 | 高 | 低 | 50%+ |
| 扫描速度 | 基准 | 3-5x | 3-5x |
| 并发性能 | GIL限制 | 真并发 | 显著 |
| 部署便利性 | 需要Python环境 | 单文件部署 | 极大提升 |

## 开发说明

### 依赖库
- `github.com/jessevdk/go-flags`: 命令行参数解析
- `go.uber.org/zap`: 高性能日志库
- `golang.org/x/text`: 文本编码处理
- `gopkg.in/yaml.v3`: YAML配置解析

### 编译要求
- Go 1.20+
- 无CGO依赖，支持交叉编译

### 构建说明
项目提供了自动化构建脚本：
- Windows: 运行 `build.bat`
- Linux: 运行 `./build.sh`

构建脚本会自动：
1. 检查Go环境
2. 下载依赖
3. 交叉编译Windows和Linux版本
4. 生成优化的二进制文件

## 贡献指南

欢迎提交Issue和Pull Request！

### 开发环境设置
```bash
git clone https://github.com/your-repo/PrivacyCheckGo.git
cd PrivacyCheckGo
go mod tidy
```

### 代码规范
- 使用`gofmt`格式化代码
- 遵循Go语言最佳实践
- 添加必要的注释和文档
- 编写单元测试

## 免责声明

本工具仅用于合法的安全测试和代码审计目的。使用者应当遵守相关法律法规，不得将本工具用于非法用途。开发者不承担因误用本工具而产生的任何责任。

## 许可证

本项目采用MIT许可证，详见[LICENSE](LICENSE)文件。

## 联系方式

如有问题或建议，请通过以下方式联系：
- 提交Issue：[GitHub Issues](../../issues)
- 邮箱：your-email@example.com

---

**PrivacyCheck Go版本 - 让代码安全检测更快更强！**
