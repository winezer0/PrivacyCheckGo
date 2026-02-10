package scanner

import (
	"fmt"
	"github.com/winezer0/xutils/logging"
	"github.com/winezer0/xutils/progress"
	"github.com/winezer0/xutils/utils"
	"privacycheck/internal/baserule"
	"sync"
)

// Scanner 扫描器
type Scanner struct {
	engine     *RuleEngine
	config     *ScanConfig // 改为使用自己的配置结构体
	cache      *CacheManager
	results    []ScanResult
	resultsMux sync.Mutex
}

// NewScanner 创建新的扫描器
func NewScanner(rules baserule.RuleMap, config *ScanConfig) (*Scanner, error) {
	engine, err := NewRuleEngine(rules)
	if err != nil {
		return nil, fmt.Errorf("创建规则引擎失败: %w", err)
	}

	scanner := &Scanner{
		engine: engine,
		config: config,
	}

	// 初始化缓存
	if config.SaveCache {
		cacheFile := config.ProjectName + ".cache"
		cache := NewCacheManager(cacheFile)
		scanner.cache = cache
	}

	return scanner, nil
}

// Scan 执行扫描
func (s *Scanner) Scan(filePaths []string) ([]ScanResult, error) {
	logging.Infof("starting scan, found %d files to process", len(filePaths))
	logging.Infof("using %d worker threads", s.config.Workers)

	// 创建工作池 - 直接使用文件路径
	jobs := make(chan string, 100) // 使用缓冲通道，避免阻塞
	results := make(chan ScanJob, 100)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < s.config.Workers; i++ {
		wg.Add(1)
		go s.worker(jobs, results, &wg)
	}

	// 发送任务 - 直接发送文件路径
	go func() {
		defer close(jobs)
		for _, filePath := range filePaths {
			jobs <- filePath
		}
	}()

	// 收集结果
	go func() {
		defer close(results)
		wg.Wait()
	}()

	// 启动进度监控
	bar := progress.NewProcessBarByTotalTask(int64(len(filePaths)), "Scanning ...")

	// 处理结果
	for job := range results {
		s.processJobResult(job)
		_ = bar.Add(1)
	}

	// 最终保存缓存
	if s.cache != nil {
		if err := s.cache.ForceSave(); err != nil {
			logging.Warnf("failed to save final cache: %v", err)
		}
	}

	return s.results, nil
}

// worker 工作协程 - 直接处理文件路径
func (s *Scanner) worker(jobs <-chan string, results chan<- ScanJob, wg *sync.WaitGroup) {
	defer wg.Done()

	for filePath := range jobs {
		job := ScanJob{FilePath: filePath}

		// 检查文件是否有效
		if !utils.FileExists(filePath) {
			job.Error = fmt.Errorf("file %s is not valid or is a directory", filePath)
			results <- job
			continue
		}

		// 检查缓存
		if s.cache != nil {
			if cachedResults, exists := s.cache.GetCachedResult(filePath); exists {
				job.Results = cachedResults
				results <- job
				continue
			}
		}

		// 执行扫描
		job.Results, job.Error = s.scanFile(filePath)

		results <- job
	}
}

// scanFile 扫描单个文件 - 直接接受文件路径
func (s *Scanner) scanFile(filePath string) ([]ScanResult, error) {
	var results []ScanResult

	// 获取文件大小和编码信息
	fileInfo, err := utils.PathToFileInfo(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file size %s: %w", filePath, err)
	}

	// 判断是否启用分块读取以及文件大小是否超过阈值
	chunkThreshold := int64(s.config.ChunkLimit) * 1024 * 1024 // 转换为字节
	if s.config.ChunkLimit > 0 && fileInfo.Size > chunkThreshold {
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

	return results, nil
}

// processJobResult 处理任务结果
func (s *Scanner) processJobResult(job ScanJob) {
	if job.Error != nil {
		logging.Warnf("failed to scan file %s: %v", job.FilePath, job.Error)
		return
	}

	// 添加结果
	s.resultsMux.Lock()
	s.results = append(s.results, job.Results...)
	s.resultsMux.Unlock()

	// 更新缓存
	if s.cache != nil {
		s.cache.SetCachedResult(job.FilePath, job.Results)

		// 定期保存缓存
		if err := s.cache.AutoSave(); err != nil {
			logging.Warnf("failed to save cache: %v", err)
		}
	}
}
