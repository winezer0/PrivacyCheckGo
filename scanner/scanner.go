package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"privacycheck/core"
	"privacycheck/logging"
	"privacycheck/utils"
)

// Scanner 扫描器
type Scanner struct {
	engine     *RuleEngine
	config     *core.Config
	cache      *ScanCache
	results    []core.ScanResult
	resultsMux sync.Mutex
	stats      core.ScanStats
	statsMux   sync.RWMutex
}

// ScanCache 扫描缓存
type ScanCache struct {
	data     core.ScanCache
	filePath string
	mux      sync.RWMutex
	lastSave time.Time
}

// NewScanner 创建新的扫描器
func NewScanner(rules map[string][]core.Rule, config *core.Config) (*Scanner, error) {
	engine, err := NewRuleEngine(rules)
	if err != nil {
		return nil, fmt.Errorf("创建规则引擎失败: %w", err)
	}

	scanner := &Scanner{
		engine: engine,
		config: config,
		stats: core.ScanStats{
			StartTime: time.Now(),
		},
	}

	// 初始化缓存
	if config.SaveCache {
		cacheFile := config.ProjectName + ".cache"
		cache, err := NewScanCache(cacheFile)
		if err != nil {
			logging.Warnf("初始化缓存失败: %v", err)
		} else {
			scanner.cache = cache
		}
	}

	return scanner, nil
}

// NewScanCache 创建新的扫描缓存
func NewScanCache(filePath string) (*ScanCache, error) {
	cache := &ScanCache{
		filePath: filePath,
		data: core.ScanCache{
			Result:     make(map[string][]core.ScanResult),
			LastUpdate: time.Now().Format(time.RFC3339),
		},
		lastSave: time.Now(),
	}

	// 尝试加载现有缓存
	if err := cache.load(); err != nil {
		logging.Warnf("加载缓存失败: %v", err)
	}

	return cache, nil
}

// load 加载缓存文件
func (c *ScanCache) load() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, err := os.Stat(c.filePath); os.IsNotExist(err) {
		return nil // 缓存文件不存在，不是错误
	}

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return fmt.Errorf("读取缓存文件失败: %w", err)
	}

	if err := json.Unmarshal(data, &c.data); err != nil {
		return fmt.Errorf("解析缓存文件失败: %w", err)
	}

	logging.Infof("加载缓存完成: 已缓存结果数: %d, 缓存更新时间: %s",
		len(c.data.Result), c.data.LastUpdate)

	return nil
}

// save 保存缓存文件
func (c *ScanCache) save() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.data.LastUpdate = time.Now().Format(time.RFC3339)

	data, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	if err := os.WriteFile(c.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入缓存文件失败: %w", err)
	}

	c.lastSave = time.Now()
	return nil
}

// get 获取缓存结果
func (c *ScanCache) get(filePath string) ([]core.ScanResult, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	results, exists := c.data.Result[filePath]
	return results, exists
}

// set 设置缓存结果
func (c *ScanCache) set(filePath string, results []core.ScanResult) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.data.Result[filePath] = results
}

// shouldSave 检查是否应该保存缓存
func (c *ScanCache) shouldSave(forceStore bool) bool {
	if forceStore {
		return true
	}
	return time.Since(c.lastSave) >= 10*time.Second // 每10秒保存一次
}

// Scan 执行扫描
func (s *Scanner) Scan(files []core.FileInfo) ([]core.ScanResult, error) {
	s.stats.TotalFiles = len(files)

	logging.Infof("开始扫描，共发现 %d 个有效文件", len(files))
	logging.Infof("使用线程数: %d", s.config.Workers)
	logging.Infof("规则引擎: %d 个规则组, %d 个规则", s.engine.GetGroupsCount(), s.engine.GetRulesCount())

	// 创建工作池
	jobs := make(chan core.FileInfo, len(files))
	results := make(chan ScanJob, len(files))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < s.config.Workers; i++ {
		wg.Add(1)
		go s.worker(jobs, results, &wg)
	}

	// 发送任务
	go func() {
		defer close(jobs)
		for _, file := range files {
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
		if err := s.cache.save(); err != nil {
			logging.Warnf("保存最终缓存失败: %v", err)
		}
	}

	s.stats.ElapsedTime = time.Since(s.stats.StartTime)
	logging.Infof("扫描完成！总用时: %v, 发现结果: %d 个",
		s.stats.ElapsedTime, len(s.results))

	return s.results, nil
}

// ScanJob 扫描任务结果
type ScanJob struct {
	FilePath string
	Results  []core.ScanResult
	Error    error
}

// worker 工作协程
func (s *Scanner) worker(jobs <-chan core.FileInfo, results chan<- ScanJob, wg *sync.WaitGroup) {
	defer wg.Done()

	for file := range jobs {
		job := ScanJob{FilePath: file.Path}

		// 检查缓存
		if s.cache != nil {
			if cachedResults, exists := s.cache.get(file.Path); exists {
				job.Results = cachedResults
				results <- job
				continue
			}
		}

		// 执行扫描
		if s.config.ChunkMode {
			job.Results, job.Error = s.scanFileInChunks(file)
		} else {
			job.Results, job.Error = s.scanFile(file)
		}

		results <- job
	}
}

// scanFile 扫描单个文件
func (s *Scanner) scanFile(file core.FileInfo) ([]core.ScanResult, error) {
	content, err := utils.ReadFileSafe(file.Path, file.Encoding)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	results := s.engine.ApplyRules(content, file.Path)
	return results, nil
}

// scanFileInChunks 分块扫描文件
func (s *Scanner) scanFileInChunks(file core.FileInfo) ([]core.ScanResult, error) {
	var allResults []core.ScanResult
	chunkSize := 1024 * 1024 // 1MB
	chunkOffset := 0

	err := utils.ReadFileInChunks(file.Path, file.Encoding, chunkSize, func(chunk string) error {
		results := s.engine.ApplyRuleToChunk(chunk, file.Path, chunkOffset)
		allResults = append(allResults, results...)
		chunkOffset += len(chunk)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("分块读取文件失败: %w", err)
	}

	return allResults, nil
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
		s.cache.set(job.FilePath, job.Results)

		// 定期保存缓存
		if s.cache.shouldSave(false) {
			if err := s.cache.save(); err != nil {
				logging.Warnf("保存缓存失败: %v", err)
			}
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

	fmt.Printf("\r当前进度: %d/%d (%.2f%%) 已用时长: %v 预计剩余: %v 发现结果: %d",
		stats.ProcessedFiles, stats.TotalFiles, percent,
		elapsed.Truncate(time.Second), remaining.Truncate(time.Second),
		stats.TotalResults)
}
