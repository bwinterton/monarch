package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// ValidateFile validates that the configuration file passed in
// has valid syntax and is complete
func ValidateFile(filename string) {

	fmt.Println("Validating config file....")

	_, err := ParseConfig(filename)
	if err != nil {
		fmt.Printf("The config file is not valid: %s\n", err)
		return
	}

	fmt.Println("Configuration file is valid!")

}

// ParseConfig parses the configuration file passed in and returns
// the completed struct with all the information found in the file
func ParseConfig(filename string) (Config, error) {

	config := Config{}

	fileConfig, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("File %s could not be opened or was not found", filename)
	}

	err = json.Unmarshal(fileConfig, &config)
	if err != nil {
		return config, fmt.Errorf("File could not be deserialized. Please check if the config is valid json")
	}

	err = validateConfig(config)
	if err != nil {
		return config, fmt.Errorf("File is not valid: %s", err)
	}

	return config, nil
}

// validateConfig validates that the Config struct passed in is complete
// and valid
func validateConfig(config Config) error {

	if config.Registry.URL == "" ||
		config.Registry.Username == "" ||
		config.Registry.Password == "" {
		return fmt.Errorf("Registry configuration is not complete. URL, username, and password are all required")
	}

	if config.Quay.Namespace == "" ||
		config.Quay.OAuthToken == "" {
		return fmt.Errorf("Quay configuration is not complete. Namespace , and OAuth Token are all required")
	}

	// TODO: Add more checking here for warnings of things like not specifying
	// new name and not specifying public explicitly. Make user aware of the defaults
	// when these things happen.

	return nil
}
