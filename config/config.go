package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Configuration struct {
	*viper.Viper
}

func Load() *Configuration {
	// Create a new configuration object
	config := &Configuration{
		Viper: viper.New(),
	}

	// Load the default configurations
	config.loadDefaultConfig()

	// Select the .env file
	config.SetConfigName(".env")
	config.SetConfigType("dotenv")
	config.AddConfigPath(".")

	// Automatically refresh environment variables
	config.AutomaticEnv()

	// Read configuration
	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("The .env file has not been found in the current directory, using default settings.")
		} else {
			fmt.Println(err.Error())
		}
	}

	// Verify the APP_PORT
	port := config.GetInt("app_port")
	if port < 1 || port > 65535 {
		panic("Invalid APP_PORT, cannot continue.")
	}

	configChanged := false

	// Verify the APP_URL
	url := config.GetString("app_url")
	if url == "" {
		// Dynamically set the APP_URL
		config.Set("app_url", "http://localhost:"+config.GetString("app_port"))
		configChanged = true
		// Print success message
		fmt.Println("The APP_URL was set dynamically")
	}

	// Verify the APP_KEY
	key := config.GetString("app_key")
	if key == "" {
		// Generate a new APP_KEY
		key, err := generateRandomStringURLSafe(32)
		if err != nil {
			panic(err)
		}
		// Set the new APP_KEY
		config.Set("app_key", key)
		configChanged = true
		// Print success message
		fmt.Println("A new APP_KEY was generated")
	}

	if configChanged {
		// Write the new configuration to the file
		err := config.SafeWriteConfig()
		if err != nil {
			fmt.Println("err in viper.SafeWriteConfig():", err)
		}

		// Rename the .env file due to a bug in viper
		// See: https://github.com/spf13/viper/issues/455
		err = os.Rename(".env.dotenv", ".env")
		if err != nil {
			fmt.Println("err in os.Rename():", err)
		}

		// Print success message
		fmt.Println("The changed were saved in the .env file")
	}

	// Return the configuration object
	return config
}

func (config *Configuration) loadDefaultConfig() {
	// App configuration
	config.SetDefault("app_name", "Fiber Boilerplate")
	config.SetDefault("app_env", "local")
	config.SetDefault("app_port", 8080)
	config.SetDefault("app_debug", true)

	// Access logger configuration
	config.SetDefault("access_logger", "file")
	config.SetDefault("access_logger_filename", "access.log")
	config.SetDefault("access_logger_maxsize", 500)
	config.SetDefault("access_logger_maxbackups", 3)
	config.SetDefault("access_logger_maxage", 28)
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateRandomStringURLSafe(n int) (string, error) {
	b, err := generateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}
