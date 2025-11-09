package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CommandLog はコマンド実行ログの構造体
type CommandLog struct {
	Timestamp  time.Time              `json:"timestamp"`
	Command    string                 `json:"command"`
	UserID     string                 `json:"user_id"`
	Username   string                 `json:"username"`
	ChannelID  string                 `json:"channel_id"`
	Success    bool                   `json:"success"`
	Error      string                 `json:"error,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ErrorLog はエラーログの構造体
type ErrorLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`  // ERROR, FATAL
	Source    string                 `json:"source"` // 関数名やモジュール名
	Message   string                 `json:"message"`
	Error     string                 `json:"error,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// CommandStats はコマンド統計の構造体
type CommandStats struct {
	TotalCommands int                    `json:"total_commands"`
	CommandCounts map[string]int         `json:"command_counts"`
	UserCounts    map[string]int         `json:"user_counts"`
	LastUpdated   time.Time              `json:"last_updated"`
	MonthlyStats  map[string]MonthlyStat `json:"monthly_stats"`
}

// MonthlyStat は月別統計の構造体
type MonthlyStat struct {
	Year          int            `json:"year"`
	Month         int            `json:"month"`
	TotalCommands int            `json:"total_commands"`
	CommandCounts map[string]int `json:"command_counts"`
	UserCounts    map[string]int `json:"user_counts"`
}

// Logger はログシステムのメイン構造体
type Logger struct {
	logDir       string
	statsFile    string
	currentMonth string
	monthlyFile  string
	errorFile    string
	stats        *CommandStats
	mutex        sync.RWMutex
}

// NewLogger は新しいロガーを作成する
func NewLogger(logDir string) *Logger {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create log directory: %v", err))
	}

	statsFile := filepath.Join(logDir, "command_stats.json")
	currentMonth := time.Now().Format("2006-01")
	monthlyFile := filepath.Join(logDir, fmt.Sprintf("commands_%s.log", currentMonth))
	errorFile := filepath.Join(logDir, fmt.Sprintf("errors_%s.log", currentMonth))

	logger := &Logger{
		logDir:       logDir,
		statsFile:    statsFile,
		currentMonth: currentMonth,
		monthlyFile:  monthlyFile,
		errorFile:    errorFile,
		stats: &CommandStats{
			TotalCommands: 0,
			CommandCounts: make(map[string]int),
			UserCounts:    make(map[string]int),
			LastUpdated:   time.Now(),
			MonthlyStats:  make(map[string]MonthlyStat),
		},
	}

	// 既存の統計を読み込み
	logger.loadStats()

	return logger
}

// LogCommand はコマンド実行をログに記録する
func (l *Logger) LogCommand(command string, userID, username, channelID string, success bool, errorMsg string, parameters map[string]interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 月が変わった場合はローテーション
	currentMonth := time.Now().Format("2006-01")
	if l.currentMonth != currentMonth {
		l.rotateLogs()
		l.currentMonth = currentMonth
	}

	// ログエントリを作成
	logEntry := CommandLog{
		Timestamp:  time.Now(),
		Command:    command,
		UserID:     userID,
		Username:   username,
		ChannelID:  channelID,
		Success:    success,
		Error:      errorMsg,
		Parameters: parameters,
	}

	// 月次ログファイルに書き込み
	l.writeToMonthlyLog(logEntry)

	// 統計を更新
	l.updateStats(logEntry)
}

// writeToMonthlyLog は月次ログファイルにログエントリを書き込む
func (l *Logger) writeToMonthlyLog(entry CommandLog) {
	file, err := os.OpenFile(l.monthlyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open monthly log file: %v\n", err)
		return
	}
	defer file.Close()

	// JSON形式でログを書き込み
	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Failed to marshal log entry: %v\n", err)
		return
	}

	file.Write(jsonData)
	file.WriteString("\n")
}

// updateStats は統計情報を更新する
func (l *Logger) updateStats(entry CommandLog) {
	// 全体統計を更新
	l.stats.TotalCommands++
	l.stats.CommandCounts[entry.Command]++
	l.stats.UserCounts[entry.UserID]++
	l.stats.LastUpdated = time.Now()

	// 月別統計を更新
	monthKey := entry.Timestamp.Format("2006-01")
	if _, exists := l.stats.MonthlyStats[monthKey]; !exists {
		l.stats.MonthlyStats[monthKey] = MonthlyStat{
			Year:          entry.Timestamp.Year(),
			Month:         int(entry.Timestamp.Month()),
			TotalCommands: 0,
			CommandCounts: make(map[string]int),
			UserCounts:    make(map[string]int),
		}
	}

	monthlyStat := l.stats.MonthlyStats[monthKey]
	monthlyStat.TotalCommands++
	monthlyStat.CommandCounts[entry.Command]++
	monthlyStat.UserCounts[entry.UserID]++
	l.stats.MonthlyStats[monthKey] = monthlyStat

	// 統計をファイルに保存
	l.saveStats()
}

// LogError はエラーをログに記録する
func (l *Logger) LogError(level, source, message string, err error, details map[string]interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 月が変わった場合はローテーション
	currentMonth := time.Now().Format("2006-01")
	if l.currentMonth != currentMonth {
		l.rotateLogs()
		l.currentMonth = currentMonth
	}

	// エラーログエントリを作成
	errorLog := ErrorLog{
		Timestamp: time.Now(),
		Level:     level,
		Source:    source,
		Message:   message,
		Details:   details,
	}

	if err != nil {
		errorLog.Error = err.Error()
	}

	// エラーログファイルに書き込み
	l.writeToErrorLog(errorLog)
}

// writeToErrorLog はエラーログファイルにログエントリを書き込む
func (l *Logger) writeToErrorLog(entry ErrorLog) {
	file, err := os.OpenFile(l.errorFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open error log file: %v\n", err)
		return
	}
	defer file.Close()

	// JSON形式でログを書き込み
	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Failed to marshal error log entry: %v\n", err)
		return
	}

	file.Write(jsonData)
	file.WriteString("\n")
}

// rotateLogs は月次ログローテーションを実行する
func (l *Logger) rotateLogs() {
	// 新しい月次ファイル名を設定
	l.monthlyFile = filepath.Join(l.logDir, fmt.Sprintf("commands_%s.log", l.currentMonth))
	l.errorFile = filepath.Join(l.logDir, fmt.Sprintf("errors_%s.log", l.currentMonth))
}

// loadStats は既存の統計ファイルを読み込む
func (l *Logger) loadStats() {
	data, err := os.ReadFile(l.statsFile)
	if err != nil {
		// ファイルが存在しない場合は新しい統計を作成
		return
	}

	var stats CommandStats
	if err := json.Unmarshal(data, &stats); err != nil {
		fmt.Printf("Failed to load stats: %v\n", err)
		return
	}

	l.stats = &stats
}

// saveStats は統計をJSONファイルに保存する
func (l *Logger) saveStats() {
	data, err := json.MarshalIndent(l.stats, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal stats: %v\n", err)
		return
	}

	if err := os.WriteFile(l.statsFile, data, 0644); err != nil {
		fmt.Printf("Failed to save stats: %v\n", err)
	}
}

// GetStats は現在の統計情報を取得する
func (l *Logger) GetStats() *CommandStats {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	// コピーを返す
	statsCopy := *l.stats
	return &statsCopy
}

// GetMonthlyLogPath は現在の月次ログファイルのパスを取得する
func (l *Logger) GetMonthlyLogPath() string {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.monthlyFile
}

// CleanupOldLogs は古いログファイルを削除する（1か月以上前）
func (l *Logger) CleanupOldLogs() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	cutoffDate := time.Now().AddDate(0, -1, 0) // 1か月前

	files, err := os.ReadDir(l.logDir)
	if err != nil {
		fmt.Printf("Failed to read log directory: %v\n", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		var dateStr string
		var isLogFile bool

		// commands_YYYY-MM.log または errors_YYYY-MM.log 形式をチェック
		if len(fileName) >= 20 && fileName[len(fileName)-4:] == ".log" {
			if fileName[:9] == "commands_" {
				dateStr = fileName[9 : len(fileName)-4] // YYYY-MM部分を抽出
				isLogFile = true
			} else if fileName[:7] == "errors_" {
				dateStr = fileName[7 : len(fileName)-4] // YYYY-MM部分を抽出
				isLogFile = true
			}
		}

		if isLogFile {
			if fileDate, err := time.Parse("2006-01", dateStr); err == nil {
				if fileDate.Before(cutoffDate) {
					filePath := filepath.Join(l.logDir, fileName)
					if err := os.Remove(filePath); err != nil {
						fmt.Printf("Failed to remove old log file %s: %v\n", filePath, err)
					} else {
						fmt.Printf("Removed old log file: %s\n", fileName)
					}
				}
			}
		}
	}
}
