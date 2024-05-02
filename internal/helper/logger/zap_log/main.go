package logger

import (
	"fmt"
	"golang-api-starter/internal/config"
	"os"
	"slices"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type OutputTypes struct {
	Console bool
	File    bool
}

const (
	DebugLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	FatalLevel
)

type ZapLog struct {
	ConsoleLogger *zap.Logger
	Output        OutputTypes
	Filename      *string
	Level         int // available level (DebugLevel,InfoLevel,WarningLevel,ErrorLevel), Debugf() in console_logger.go will be effective if DebugLevel is set
	DebugSymbol   *string
	mu            sync.Mutex
}

var cfg = config.Cfg
var Zlog = &ZapLog{}

// Set up lumberjack as a logger:
var logFile = &lumberjack.Logger{
	// Filename:   fmt.Sprintf("./log/%s", "requests.log"), // Or any other path
	MaxSize:    500,  // MB; after this size, a new log file is created
	MaxBackups: 10,   // Number of backups to keep
	MaxAge:     28,   // Days
	Compress:   true, // Compress the backups using gzip
}

func NewZlog() {
	Zlog.Output = OutputTypes{
		File:    slices.Contains(cfg.Logging.Zap.Output, "file"),
		Console: slices.Contains(cfg.Logging.Zap.Output, "console"),
	}
	Zlog.Level = cfg.Logging.Level
	Zlog.DebugSymbol = cfg.Logging.DebugSymbol
	Zlog.Filename = &cfg.Logging.Zap.Filename
	logFile.Filename = fmt.Sprintf("./log/%s", *Zlog.Filename)
}

func (zl *ZapLog) SetDebugSymbol(symbol *string) *ZapLog {
	zl.DebugSymbol = symbol
	return zl
}

func GetField(key string, value interface{}) zap.Field {
	return Zlog.getField(key, value)
}
func (zl *ZapLog) getField(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

func SetLevel(lvl int) *ZapLog {
	return Zlog.setLevel(lvl)
}
func (zl *ZapLog) setLevel(lvl int) *ZapLog {
	zl.Level = lvl
	return zl
}

// SetFilename sets the custom filename under log/ directory
func (zl *ZapLog) SetFilename(filename string) {
	*zl.Filename = filename
}

// SysLog output the message into file and console depends on config
// SetLevel can override the default Zlog.Level before calling this func
func SysLog(msg string, keysAndValues ...zapcore.Field) {
	Zlog.sysLog(msg, keysAndValues...)
}
func (zl *ZapLog) sysLog(msg string, keysAndValues ...zapcore.Field) {
	zl.mu.Lock()
	if !Zlog.Output.Console && !Zlog.Output.File {
		return
	}
	logger, err := fileLogger(*zl.Filename, zl.Output)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}
	defer logger.Sync() // Ensure logs are flushed

	if zl.Level <= InfoLevel {
		logger.Info(msg, keysAndValues...)
	} else if zl.Level <= WarningLevel {
		logger.Warn(msg, keysAndValues...)
	} else if zl.Level <= ErrorLevel {
		logger.Error(msg, keysAndValues...)
	} else if zl.Level <= FatalLevel {
		logger.Fatal(msg, keysAndValues...)
	}

	zl.setLevel(cfg.Logging.Level) // reset to config's defined level
	zl.mu.Unlock()
}

// fileLogger initializes a zap.Logger that writes to both the console and a specified file.
// ref: https://www.golinuxcloud.com/golang-zap-logger/#Setting_Output_in_Zap_Console_Log_File_or_Both
func fileLogger(filename string, outputTypes OutputTypes) (*zap.Logger, error) {
	// Configure the time format
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.RFC3339TimeEncoder

	// Create file and console encoders
	fileEncoder := zapcore.NewJSONEncoder(config)
	consoleEncoder := zapcore.NewConsoleEncoder(config)

	// Open the log file
	// Default logFile without log rotate
	// logFilePath := fmt.Sprintf("./log/%s", filename)
	// logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to open log file: %v", err)
	// }

	// Create writers for file and console
	fileWriter := zapcore.AddSync(logFile)
	consoleWriter := zapcore.AddSync(os.Stdout)

	// Set the log level
	defaultLogLevel := zapcore.DebugLevel

	// Create cores for writing to the file and console
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, defaultLogLevel)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, defaultLogLevel)

	// Combine cores
	var core zapcore.Core
	if outputTypes.Console && outputTypes.File {
		core = zapcore.NewTee(fileCore, consoleCore)
	} else if outputTypes.Console && !outputTypes.File {
		core = zapcore.NewTee(consoleCore)
	} else if !outputTypes.Console && outputTypes.File {
		core = zapcore.NewTee(fileCore)
	}

	// Create the logger with additional context information (caller, stack trace)
	// logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))

	return logger, nil
}
