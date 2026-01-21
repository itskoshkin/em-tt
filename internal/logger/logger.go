package logger

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"subscription-aggregator-service/internal/config"
	"subscription-aggregator-service/internal/utils/request"
)

func SetupLogger() {
	fmt.Print("Setting up logger... ")

	if !viper.GetBool(config.LogEnabled) {
		return
	}

	var writers []io.Writer
	writers = append(writers, os.Stdout)

	if viper.GetBool(config.LogToFile) {
		file, err := os.OpenFile(viper.GetString(config.LogFilePath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("")
			log.Printf("Failed to open log file, will fallback to stdout: %v", err)
		} else {
			writers = append(writers, file)
		}
	}

	level := slog.LevelWarn
	switch viper.GetString(config.LogLevel) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	}

	switch viper.GetString(config.LogFormat) {
	case "text":
		slog.SetDefault(slog.New(slog.NewTextHandler(io.MultiWriter(writers...), &slog.HandlerOptions{Level: level})))
	case "json":
		slog.SetDefault(slog.New(slog.NewTextHandler(io.MultiWriter(writers...), &slog.HandlerOptions{Level: level})))
	default:
		fmt.Println("")
		log.Printf("Unknown log format (\"%s\"), will fallback to text format\n", viper.GetString(config.LogFormat))
		slog.SetDefault(slog.New(slog.NewTextHandler(io.MultiWriter(writers...), &slog.HandlerOptions{Level: level})))
	}

	fmt.Println(" Done.")
}

func GinLoggerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		path := ctx.Request.URL.Path
		query := ctx.Request.URL.RawQuery
		method := ctx.Request.Method

		ctx.Next() // Blocks until request is processed

		duration := time.Since(start)
		status := ctx.Writer.Status()
		errors := ctx.Errors.String()
		ipAddr := ctx.ClientIP()
		userAgent := ctx.Request.UserAgent()

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		args := []any{
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", ipAddr),
			slog.String("user_agent", userAgent),
			slog.Duration("duration", duration),
		}

		if id, ok := request.FromContext(ctx.Request.Context()); ok {
			args = append(args, slog.String("request_id", id))
		}

		if errors != "" {
			args = append(args, slog.String("errors", errors))
		}

		slog.Log(ctx.Request.Context(), level, "http_request", args...)
	}
}
