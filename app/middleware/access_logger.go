package middleware

import (
	"github.com/thomasvvugt/fiber-boilerplate/config"

	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func AccessLogger(config *config.Configuration) fiber.Handler {
	var logger *zap.Logger
	// Use the access.log file, the console or turn the logger off
	switch strings.ToLower(config.GetString("access_logger")) {
	case "file":
		// Create an access.log file using lumberjack for zap
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.GetString("access_logger_filename"),
			MaxSize:    config.GetInt("access_logger_maxsize"), // in megabytes
			MaxBackups: config.GetInt("access_logger_maxbackups"),
			MaxAge:     config.GetInt("access_logger_maxage"), // in days
		})
		// Create a zap core object for JSON encoding
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			w,
			zap.InfoLevel,
		)
		// Create a zap logger object
		logger = zap.New(core)
		break
	case "off":
		// Return an empty function to disable access logging
		return func(c *fiber.Ctx) {}
	default:
		// Create a zap logger object
		var err error
		if config.GetString("app_env") == "production" {
			logger, err = zap.NewProduction()
		} else {
			logger, err = zap.NewDevelopment()
		}
		if err != nil {
			fmt.Println(err)
		}
		break
	}

	// Flush buffers, if any
	defer logger.Sync()

	// Return the access logger middleware function
	return func(c *fiber.Ctx) {
		// Handle the request to calculate the number of bytes sent
		c.Next()

		// Send structured information message to the logger
		logger.Info(c.IP()+" - "+c.Method()+" "+c.OriginalURL()+" - "+strconv.Itoa(c.Fasthttp.Response.StatusCode())+
			" - "+strconv.Itoa(len(c.Fasthttp.Response.Body())),

			zap.String("ip", c.IP()),
			zap.String("hostname", c.Hostname()),
			zap.String("method", c.Method()),
			zap.String("path", c.OriginalURL()),
			zap.String("protocol", c.Protocol()),
			zap.Int("status", c.Fasthttp.Response.StatusCode()),

			zap.String("x-forwarded-for", c.Get(fiber.HeaderXForwardedFor)),
			zap.String("user-agent", c.Get(fiber.HeaderUserAgent)),
			zap.String("referer", c.Get(fiber.HeaderReferer)),

			zap.Int("bytesReceived", len(c.Fasthttp.Request.Body())),
			zap.Int("bytesSent", len(c.Fasthttp.Response.Body())),
		)
	}
}
