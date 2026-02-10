package scanner

import (
	"fmt"
	"github.com/winezer0/xutils/cacher"
	"github.com/winezer0/xutils/logging"
	"github.com/winezer0/xutils/progress"
	"github.com/winezer0/xutils/utils"
	"privacycheck/internal/baserule"
	"sync"
)

// Scanner 扫描器
type Scanner struct {
	workers      int
	chunkLimit   int
	engine       *RuleEngine
	cacheManager *cacher.CacheManager
}

// NewScanner 创建新的扫描器
func NewScanner(rules baserule.RuleMap, config *ScanConfig) (*Scanner, error) {
	engine, err := NewRuleEngine(rules)
	if err != nil {
		return nil, fmt.Errorf("创建规则引擎失败: %w", err)
	}

	scanner := &Scanner{
		engine:       engine,
		workers:      config.Workers,
		chunkLimit:   config.ChunkLimit,
		cacheManager: cacher.NewCacheManager(config.CacheFile),
	}

	return scanner, nil
}

func (s *Scanner) Scan(filePaths []string) ([]ScanResult, error) {
	logging.Infof("starting scan files: %d worker: %d", len(filePaths), s.workers)
	bar := progress.NewProcessBarByTotalTask(int64(len(filePaths)), "Scanning ...")
	scanJobs := make(chan string, 100)
	scanResults := make(chan ScanJob, 100) // 缓冲足够容纳所有结果
	var wg sync.WaitGroup
	// 启动 workers（内联或保留 worker 函数均可）
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go s.worker(scanJobs, scanResults, &wg)
	}

	// 发送任务
	go func() {
		defer close(scanJobs)
		for _, filePath := range filePaths {
			scanJobs <- filePath
		}
	}()

	// 等待 workers 退出（可选，主要用于资源清理，不影响结果收集）
	go func() {
		wg.Wait()
		// 此处不再关闭 scanResults！
	}()

	var allResults []ScanResult
	var resultsMux sync.Mutex

	// 收集 exactly N 个结果
	for i := 0; i < len(filePaths); i++ {
		job := <-scanResults
		_ = bar.Add(1)
		if job.Error != nil {
			logging.Warnf("failed to scan file %s: %v", job.FilePath, job.Error)
		} else {
			resultsMux.Lock()
			allResults = append(allResults, job.Results...)
			resultsMux.Unlock()
		}
	}

	s.cacheManager.Clear()
	return allResults, nil
}

// worker 工作协程 - 直接处理文件路径
func (s *Scanner) worker(jobs <-chan string, results chan<- ScanJob, wg *sync.WaitGroup) {
	defer wg.Done()
	for filePath := range jobs {
		// 执行扫描
		job := ScanJob{FilePath: filePath}
		job.Results, job.Error = s.scanFile(filePath)
		results <- job
	}
}

// scanFile 扫描单个文件 - 直接接受文件路径
func (s *Scanner) scanFile(filePath string) ([]ScanResult, error) {
	// 检查缓存
	var cachedResults []ScanResult
	if ok := s.cacheManager.GetAs(filePath, &cachedResults); ok {
		return cachedResults, nil
	}

	// 存储扫描结果
	var results []ScanResult
	// 获取文件大小和编码信息
	fileInfo, err := utils.PathToFileInfo(filePath)
	if err != nil || fileInfo.Size == 0 {
		return nil, fmt.Errorf("failed to get file info %s: %w", filePath, err)
	}

	// 判断是否启用分块读取以及文件大小是否超过阈值
	chunkThreshold := int64(s.chunkLimit) * 1024 * 1024 // 转换为字节
	if s.chunkLimit > 0 && fileInfo.Size > chunkThreshold {
		const chunkSize = 1024 * 1024 // 1MB per chunk
		err := utils.ReadFileByChunk(filePath, fileInfo.Encoding, chunkSize, func(chunk utils.ChunkInfo) error {
			// 对每个块应用规则，传入正确的行号偏移
			chunkResults := s.engine.ApplyRules(chunk.Content, filePath, int(chunk.StartOffset), chunk.StartLine)
			results = append(results, chunkResults...)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to read the large file %s error: %w", filePath, err)
		}
	} else {
		// 小文件或禁用分块读取时，直接读取全部内容
		content, err := utils.ReadFileWithEncoding(filePath, fileInfo.Encoding)
		if err != nil {
			return nil, fmt.Errorf("failed to read the file %s error: %w", filePath, err)
		}
		results = s.engine.ApplyRules(content, filePath, 0, 1)
	}

	// 更新缓存
	s.cacheManager.Set(filePath, results)
	return results, nil
}
