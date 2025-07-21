package scanner

import (
	"fmt"
	"privacycheck/internal/baserule"
	"privacycheck/pkg/fileutils"
	"privacycheck/pkg/logging"
	"sync"
	"time"
)

// Scanner 扫描器
type Scanner struct {
	engine     *RuleEngine
	config     *ScanConfig // 改为使用自己的配置结构体
	cache      *CacheManager
	results    []ScanResult
	resultsMux sync.Mutex
	stats      ScanStats
	statsMux   sync.RWMutex
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
		stats: ScanStats{
			StartTime: time.Now(),
		},
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
	// 转换文件路径为FileInfo
	fileInfos, err := fileutils.ConvertPathsToInfos(filePaths, s.config.Workers)
	if len(fileInfos) == 0 {
		logging.Warnf("获取文件信息失败 Error: %v", err)
	}
	s.stats.TotalFiles = len(fileInfos)

	logging.Infof("开始扫描，共发现 %d 个有效文件", len(fileInfos))
	logging.Infof("使用线程数: %d", s.config.Workers)

	// 创建工作池
	jobs := make(chan fileutils.FileInfo, len(fileInfos))
	results := make(chan ScanJob, len(fileInfos))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < s.config.Workers; i++ {
		wg.Add(1)
		go s.worker(jobs, results, &wg)
	}

	// 发送任务
	go func() {
		defer close(jobs)
		for _, file := range fileInfos {
			jobs <- file
		}
	}()

	// 启动进度监控
	progressDone := make(chan bool)
	go s.progressMonitor(progressDone)

	// 收集结果
	go func() {
		defer close(results)
		wg.Wait()
	}()

	// 处理结果
	for job := range results {
		s.processJobResult(job)
	}

	// 停止进度监控
	progressDone <- true

	// 最终保存缓存
	if s.cache != nil {
		if err := s.cache.ForceSave(); err != nil {
			logging.Warnf("保存最终缓存失败: %v", err)
		}
	}

	s.stats.ElapsedTime = time.Since(s.stats.StartTime)
	logging.Infof("扫描完成！总用时: %v, 发现结果: %d 个",
		s.stats.ElapsedTime, len(s.results))

	return s.results, nil
}

// worker 工作协程
func (s *Scanner) worker(jobs <-chan fileutils.FileInfo, results chan<- ScanJob, wg *sync.WaitGroup) {
	defer wg.Done()

	for file := range jobs {
		job := ScanJob{FilePath: file.Path}

		// 检查缓存
		if s.cache != nil {
			if cachedResults, exists := s.cache.GetCachedResult(file.Path); exists {
				job.Results = cachedResults
				results <- job
				continue
			}
		}

		// 执行扫描
		job.Results, job.Error = s.scanFile(file)

		results <- job
	}
}

// scanFile 扫描单个文件
func (s *Scanner) scanFile(fileInfo fileutils.FileInfo) ([]ScanResult, error) {
	var results []ScanResult

	// 判断是否启用分块读取以及文件大小是否超过阈值
	chunkThreshold := int64(s.config.ChunkLimit) * 1024 * 1024 // 转换为字节
	if s.config.ChunkLimit > 0 && fileInfo.Size > chunkThreshold {
		const chunkSize = 1024 * 1024 // 1MB per chunk
		err := fileutils.ReadFileByChunk(fileInfo.Path, fileInfo.Encoding, chunkSize, func(chunk fileutils.ChunkInfo) error {
			// 对每个块应用规则，传入正确的行号偏移
			chunkResults := s.engine.ApplyRules(chunk.Content, fileInfo.Path, int(chunk.StartOffset), chunk.StartLine)
			results = append(results, chunkResults...)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to read the large file %s error: %w", fileInfo.Path, err)
		}
	} else {
		// 小文件或禁用分块读取时，直接读取全部内容
		content, err := fileutils.ReadFileWithEncoding(fileInfo.Path, fileInfo.Encoding)
		if err != nil {
			return nil, fmt.Errorf("Failed to read the file %s error: %w", fileInfo.Path, err)
		}
		results = s.engine.ApplyRules(content, fileInfo.Path, 0, 1)
	}

	return results, nil
}

// processJobResult 处理任务结果
func (s *Scanner) processJobResult(job ScanJob) {
	s.statsMux.Lock()
	s.stats.ProcessedFiles++
	s.statsMux.Unlock()

	if job.Error != nil {
		logging.Warnf("扫描文件失败 %s: %v", job.FilePath, job.Error)
		return
	}

	// 添加结果
	s.resultsMux.Lock()
	s.results = append(s.results, job.Results...)
	s.stats.TotalResults = len(s.results)
	s.resultsMux.Unlock()

	// 更新缓存
	if s.cache != nil {
		s.cache.SetCachedResult(job.FilePath, job.Results)

		// 定期保存缓存
		if err := s.cache.AutoSave(); err != nil {
			logging.Warnf("保存缓存失败: %v", err)
		}
	}
}

// progressMonitor 进度监控
func (s *Scanner) progressMonitor(done <-chan bool) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			s.printProgress()
		}
	}
}

// printProgress 打印进度
func (s *Scanner) printProgress() {
	s.statsMux.RLock()
	stats := s.stats
	s.statsMux.RUnlock()

	if stats.TotalFiles == 0 {
		return
	}

	elapsed := time.Since(stats.StartTime)
	percent := float64(stats.ProcessedFiles) / float64(stats.TotalFiles) * 100

	var remaining time.Duration
	if stats.ProcessedFiles > 0 {
		avgTime := elapsed / time.Duration(stats.ProcessedFiles)
		remaining = avgTime * time.Duration(stats.TotalFiles-stats.ProcessedFiles)
	}

	logging.Infof("当前进度: %d/%d (%.2f%%) 已用时长: %v 预计剩余: %v 发现结果: %d",
		stats.ProcessedFiles, stats.TotalFiles, percent,
		elapsed.Truncate(time.Second), remaining.Truncate(time.Second),
		stats.TotalResults)
}

// GetStats 获取扫描统计信息
func (s *Scanner) GetStats() *ScanStats {
	s.statsMux.RLock()
	defer s.statsMux.RUnlock()
	return &s.stats
}
