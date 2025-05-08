package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/xuri/excelize/v2"
)

// 网站记录结构体
type WebsiteRecord struct {
	MediaURL    string
	MatchResult int
	RowIndex    int
	Error       string // 记录错误信息
}

// 关键词列表
var keywords = []string{"关键词1", "关键词2"}

// 并发数量控制
const maxConcurrent = 5
const requestTimeout = 30 * time.Second
const requestInterval = 1 * time.Second

func main() {
	startTime := time.Now()

	// 设置输入和输出文件名
	inputFile := "网站列表_样例数据.xlsx"
	outputFile := "网站列表_样例数据_结果_" + time.Now().Format("20060102150405") + ".xlsx"

	// 1. 读取Excel文件
	records, err := readExcelFile(inputFile)
	if err != nil {
		log.Fatalf("读取Excel文件失败: %v", err)
	}

	fmt.Printf("共读取 %d 条网站记录\n", len(records))

	// 2 & 3. 分析每个网站（使用有限并发）
	processWebsites(records)

	// 4. 导出结果为Excel文件
	err = exportToExcel(inputFile, outputFile, records)
	if err != nil {
		log.Fatalf("导出Excel文件失败: %v", err)
	}

	fmt.Printf("分析完成，结果已导出至 %s\n", outputFile)
	fmt.Printf("总耗时: %v\n", time.Since(startTime))
}

// 读取Excel文件
func readExcelFile(filePath string) ([]WebsiteRecord, error) {
	var records []WebsiteRecord

	fmt.Println("开始读取Excel文件...")
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开Excel文件失败: %w", err)
	}
	defer f.Close()

	// 获取所有工作表名
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("Excel文件中没有工作表")
	}

	// 使用第一个工作表
	sheetName := sheetList[0]
	fmt.Printf("使用工作表: %s\n", sheetName)
	
	// 获取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("工作表为空")
	}

	// 找到media_url列的索引
	var mediaURLColIndex = -1
	if len(rows) > 0 {
		for i, cell := range rows[0] {
			if strings.ToLower(cell) == "media_url" {
				mediaURLColIndex = i
				break
			}
		}
	}

	if mediaURLColIndex == -1 {
		return nil, fmt.Errorf("未找到'media_url'列")
	}

	fmt.Printf("找到'media_url'列，索引为: %d\n", mediaURLColIndex)

	// 从第2行开始读取数据（跳过表头）
	validURLCount := 0
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) <= mediaURLColIndex {
			continue // 跳过没有足够列的行
		}

		mediaURL := row[mediaURLColIndex]
		mediaURL = strings.TrimSpace(mediaURL)
		if mediaURL == "" {
			continue // 跳过空URL
		}

		validURLCount++
		// 将数据添加到记录中
		records = append(records, WebsiteRecord{
			MediaURL: mediaURL,
			RowIndex: i + 1, // Excel行索引从1开始，且需要考虑表头
		})
	}

	fmt.Printf("有效URL数量: %d\n", validURLCount)
	return records, nil
}

// 使用有限并发处理网站
func processWebsites(records []WebsiteRecord) {
	fmt.Printf("开始处理网站数据，并发数: %d\n", maxConcurrent)
	
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	
	for i := range records {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量
		
		go func(i int) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量
			
			record := &records[i]
			fmt.Printf("[%d/%d] 正在分析网站: %s\n", i+1, len(records), record.MediaURL)
			
			// 获取网站源代码并检查关键词
			html, err := fetchWebsite(record.MediaURL)
			if err != nil {
				record.Error = err.Error()
				fmt.Printf("获取网站源代码失败: %v\n", err)
				record.MatchResult = 0
				return
			}
			
			// 检查网站源代码中是否包含关键词
			if containsKeywords(html) {
				record.MatchResult = 1
				fmt.Printf("[%d/%d] 关键词匹配成功: %s\n", i+1, len(records), record.MediaURL)
			} else {
				record.MatchResult = 0
				fmt.Printf("[%d/%d] 关键词匹配失败: %s\n", i+1, len(records), record.MediaURL)
			}
			
			// 避免请求过于频繁
			time.Sleep(requestInterval)
		}(i)
	}
	
	wg.Wait()
	fmt.Println("所有网站处理完成")
}

// 获取网站源代码
func fetchWebsite(url string) (string, error) {
	// 如果URL不包含协议，添加http://
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	// 创建上下文，设置超时
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	
	// 设置User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// 设置客户端，添加超时设置
	client := &http.Client{
		Timeout: requestTimeout,
	}

	// 发送GET请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("状态码错误: %d", resp.StatusCode)
	}

	// 使用goquery解析HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("解析HTML失败: %w", err)
	}

	// 获取HTML内容
	html, err := doc.Html()
	if err != nil {
		return "", fmt.Errorf("获取HTML内容失败: %w", err)
	}

	return html, nil
}

// 检查HTML中是否包含关键词
func containsKeywords(html string) bool {
	// 转换为小写以进行不区分大小写的搜索
	htmlLower := strings.ToLower(html)
	
	for _, keyword := range keywords {
		if strings.Contains(htmlLower, strings.ToLower(keyword)) {
			return true
		}
	}
	
	return false
}

// 导出结果为Excel文件
func exportToExcel(inputFile, outputFile string, records []WebsiteRecord) error {
	fmt.Println("开始导出结果到Excel文件...")
	
	// 打开原始Excel文件
	f, err := excelize.OpenFile(inputFile)
	if err != nil {
		return fmt.Errorf("打开Excel文件失败: %w", err)
	}
	defer f.Close()

	// 获取第一个工作表名
	sheetName := f.GetSheetList()[0]
	
	// 找到最后一列的索引
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("读取工作表失败: %w", err)
	}
	
	// 添加match_result列
	lastColIndex := len(rows[0])
	matchColName, err := excelize.ColumnNumberToName(lastColIndex + 1)
	if err != nil {
		return fmt.Errorf("列索引转换失败: %w", err)
	}
	
	// 添加error_msg列
	errorColName, err := excelize.ColumnNumberToName(lastColIndex + 2)
	if err != nil {
		return fmt.Errorf("列索引转换失败: %w", err)
	}
	
	// 设置新列的标题
	f.SetCellValue(sheetName, matchColName+"1", "match_result")
	f.SetCellValue(sheetName, errorColName+"1", "error_msg")
	
	// 设置列宽度
	f.SetColWidth(sheetName, matchColName, matchColName, 15)
	f.SetColWidth(sheetName, errorColName, errorColName, 30)
	
	// 填充匹配结果
	successCount := 0
	failCount := 0
	errorCount := 0
	
	for _, record := range records {
		// 设置match_result
		matchCell := matchColName + fmt.Sprintf("%d", record.RowIndex)
		f.SetCellValue(sheetName, matchCell, record.MatchResult)
		
		// 设置error_msg
		errorCell := errorColName + fmt.Sprintf("%d", record.RowIndex)
		f.SetCellValue(sheetName, errorCell, record.Error)
		
		// 统计结果
		if record.Error != "" {
			errorCount++
		} else if record.MatchResult == 1 {
			successCount++
		} else {
			failCount++
		}
	}
	
	// 保存到新文件
	if err := f.SaveAs(outputFile); err != nil {
		return fmt.Errorf("保存Excel文件失败: %w", err)
	}
	
	fmt.Printf("结果统计:\n")
	fmt.Printf("- 匹配成功: %d\n", successCount)
	fmt.Printf("- 匹配失败: %d\n", failCount)
	fmt.Printf("- 处理错误: %d\n", errorCount)
	fmt.Printf("- 总记录数: %d\n", len(records))
	
	return nil
} 