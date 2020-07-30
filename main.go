package main

import (
	boilerplate "github.com/thomasvvugt/fiber-boilerplate/app"
	configuration "github.com/thomasvvugt/fiber-boilerplate/config"

	"github.com/gofiber/fiber"
)

func main() {
	// Load the required configuration for the application
	config := configuration.Load()

	// Create a new application
	app := boilerplate.New(config)

	// Do stuff

	app.Get("/", func(c *fiber.Ctx) {
		c.SendString("Hello, World!")
	})

	app.Get("/panic", func(c *fiber.Ctx) {
		panic(&fiber.Error{
			Code:    500,
			Message: "Panic test!",
		})
	})

	// Start listening
	app.Listen()

	// Gracefully shutdown the application
	app.Shutdown()
}
