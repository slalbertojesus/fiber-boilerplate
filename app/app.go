package app

import (
	"github.com/thomasvvugt/fiber-boilerplate/app/middleware"
	"github.com/thomasvvugt/fiber-boilerplate/config"

	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber"
	fibermw "github.com/gofiber/fiber/middleware"
)

type App struct {
	*fiber.App
	mutex sync.Mutex
	// App settings
	Settings *Settings
	// App configuration
	Config *config.Configuration
}

type Settings struct {
	*fiber.Settings
	DisableStatusMessages bool
}

func New(config *config.Configuration) *App {
	// Use provided configuration
	settings := &Settings{
		Settings: &fiber.Settings{
			ErrorHandler:          errorHandler,
			ServerHeader:          "fiber",
			ETag:                  true,
			Prefork:               config.GetString("app_env") == "production",
			DisableStartupMessage: true,
			ReadTimeout:           120 * time.Second,
			WriteTimeout:          120 * time.Second,
			IdleTimeout:           150 * time.Second,
		},
	}

	// Create an application object
	app := App{
		App: fiber.New(settings.Settings),
		// Set app settings
		Settings: settings,
		// Set provided configuration
		Config: config,
	}

	// Use the official Recover middleware
	app.Use(fibermw.Recover())

	// Use the Access Logger middleware
	app.Use(middleware.AccessLogger(config))

	// Handle interrupt signals for clean shutdown
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		app.Shutdown()
	}()

	return &app
}

func errorHandler(ctx *fiber.Ctx, err error) {
	// Status code defaults to 500
	code := fiber.StatusInternalServerError

	// Set error message
	message := err.Error()

	// Check if it's a fiber.Error type
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Return HTTP response
	ctx.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
	// TODO: Add 500 templating and use the line below on templating error
	ctx.Status(code).SendString(message)
}

func (app *App) Shutdown() {
	// Only call once
	app.mutex.Lock()
	defer app.mutex.Unlock()

	if !app.Settings.DisableStatusMessages {
		fmt.Println("Shutting down", app.Config.GetString("app_name")+"...")
	}

	// Shutdown Fiber
	err := app.App.Shutdown()
	if err != nil {
		fmt.Println("err on Shutdown():", err)
	}

	if !app.Settings.DisableStatusMessages {
		fmt.Println(app.Config.GetString("app_name"), "was shutdown")
	}

	// Exit the application
	os.Exit(0)
}

func (app *App) Listen(tlsconfig ...*tls.Config) {
	var err error

	if !app.Settings.DisableStatusMessages {
		fmt.Println(app.Config.GetString("app_name"), "started listening on port", app.Config.GetString("app_port"))
	}

	// Start listening with or without TLS configuration
	if len(tlsconfig) > 0 {
		err = app.App.Listen(app.Config.GetInt("app_port"), tlsconfig[0])
	} else {
		err = app.App.Listen(app.Config.GetInt("app_port"))
	}

	if err != nil {
		// Listener error -> net.Listen(network, addr)
		// Permanent error when accepting new connections -> FastHTTP acceptConn()
		fmt.Println("err on Listen():", err)
	}
}
