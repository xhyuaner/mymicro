package log

import (
	"context"
	stdlog "log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogHelper interface {
	Info(msg string, fields ...Field)
	Infof(format string, v ...interface{})
	Warn(msg string, fields ...Field)
	Warnf(format string, v ...interface{})
	Error(msg string, fields ...Field)
	Errorf(format string, v ...interface{})
}

// Init initializes logger with specified options.
func Init(opts *Options) {
	mu.Lock()
	defer mu.Unlock()
	std = New(opts)
}

// New create logger by opts which can custmoized by command arguments.
func New(opts *Options) *Logger {
	if opts == nil {
		opts = NewOptions()
	}

	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(opts.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}
	encodeLevel := zapcore.CapitalLevelEncoder
	// when output to local path, with color is forbidden
	if opts.Format == consoleFormat && opts.EnableColor {
		encodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "timestamp",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    encodeLevel,
		EncodeTime:     timeEncoder,
		EncodeDuration: milliSecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	loggerConfig := &zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       opts.Development,
		DisableCaller:     opts.DisableCaller,
		DisableStacktrace: opts.DisableStacktrace,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         opts.Format,
		EncoderConfig:    encoderConfig,
		OutputPaths:      opts.OutputPaths,
		ErrorOutputPaths: opts.ErrorOutputPaths,
	}

	var err error
	l, err := loggerConfig.Build(zap.AddStacktrace(zapcore.PanicLevel), zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	logger := &Logger{
		Logger: l,

		//有时我们稍微封装了一下记录日志的方法，但是我们希望输出的文件名和行号是调用封装函数的位置。这时可以使用zap.AddCallerSkip(skip int)向上跳 1 层：
		skipCaller: l.WithOptions(zap.AddCallerSkip(1)),

		minLevel:         zapLevel,
		errorStatusLevel: zap.ErrorLevel,
		caller:           true,
		withTraceID:      true,
		//stackTrace:       true,
	}
	zap.RedirectStdLog(l)

	return logger
}

// ZapLogger used for other log wrapper such as klog.
func ZapLogger() *zap.Logger {
	return std.Logger
}

// CheckIntLevel used for other log wrapper such as klog which return if logging a
// message at the specified level is enabled.
func CheckIntLevel(level int32) bool {
	var lvl zapcore.Level
	if level < 5 {
		lvl = zapcore.InfoLevel
	} else {
		lvl = zapcore.DebugLevel
	}
	checkEntry := std.Logger.Check(lvl, "")

	return checkEntry != nil
}

// Debug method output debug level log.
func Debug(msg string, fields ...Field) {
	std.Logger.Debug(msg, fields...)
}

// Debug method output debug level log.
func DebugC(ctx context.Context, msg string, fields ...Field) {
	std.DebugContext(ctx, msg, fields...)
}

// Debugf method output debug level log.
func Debugf(format string, v ...interface{}) {
	std.Logger.Sugar().Debugf(format, v...)
}

// Debugf method output debug level log.
func DebugfC(ctx context.Context, format string, v ...interface{}) {
	std.DebugfContext(ctx, format, v...)
}

// Debugw method output debug level log.
func Debugw(msg string, keysAndValues ...interface{}) {
	std.Logger.Sugar().Debugw(msg, keysAndValues...)
}

func DebugwC(ctx context.Context, msg string, keysAndValues ...interface{}) {
	std.DebugfContext(ctx, msg, keysAndValues...)
}

// Info method output info level log.
func Info(msg string, fields ...Field) {
	std.Logger.Info(msg, fields...)
}

func InfoC(ctx context.Context, msg string, fields ...Field) {
	std.InfoContext(ctx, msg, fields...)
}

// Infof method output info level log.
func Infof(format string, v ...interface{}) {
	std.Logger.Sugar().Infof(format, v...)
}

func InfofC(ctx context.Context, format string, v ...interface{}) {
	std.InfofContext(ctx, format, v...)
}

// Warn method output warning level log.
func Warn(msg string, fields ...Field) {
	std.Logger.Warn(msg, fields...)
}

func WarnC(ctx context.Context, msg string, fields ...Field) {
	std.WarnContext(ctx, msg, fields...)
}

// Warnf method output warning level log.
func Warnf(format string, v ...interface{}) {
	std.Logger.Sugar().Warnf(format, v...)
}

func WarnfC(ctx context.Context, format string, v ...interface{}) {
	std.WarnfContext(ctx, format, v...)
}

// Error method output error level log.
func Error(msg string, fields ...Field) {
	std.Logger.Error(msg, fields...)
}

func ErrorC(ctx context.Context, msg string, fields ...Field) {
	std.ErrorContext(ctx, msg, fields...)
}

// Errorf method output error level log.
func Errorf(format string, v ...interface{}) {
	std.Logger.Sugar().Errorf(format, v...)
}

func ErrorfC(ctx context.Context, format string, v ...interface{}) {
	std.ErrorfContext(ctx, format, v...)
}

// Panic method output panic level log and shutdown application.
func Panic(msg string, fields ...Field) {
	std.Logger.Panic(msg, fields...)
}

func PanicC(ctx context.Context, msg string, fields ...Field) {
	std.PanicContext(ctx, msg, fields...)
}

// Panicf method output panic level log and shutdown application.
func Panicf(format string, v ...interface{}) {
	std.Logger.Sugar().Panicf(format, v...)
}

func PanicfC(ctx context.Context, format string, v ...interface{}) {
	std.PanicfContext(ctx, format, v...)
}

// Fatal method output fatal level log.
func Fatal(msg string, fields ...Field) {
	std.Logger.Fatal(msg, fields...)
}

func FatalC(ctx context.Context, msg string, fields ...Field) {
	std.PanicContext(ctx, msg, fields...)
}

// Fatalf method output fatal level log.
func Fatalf(format string, v ...interface{}) {
	std.Logger.Sugar().Fatalf(format, v...)
}

func FatalfC(ctx context.Context, format string, v ...interface{}) {
	std.FatalfContext(ctx, format, v...)
}

func StdInfoLogger() *stdlog.Logger {
	if std == nil {
		return nil
	}
	if l, err := zap.NewStdLogAt(std.Logger, zapcore.InfoLevel); err == nil {
		return l
	}

	return nil
}

func Flush() { std.Flush() }
