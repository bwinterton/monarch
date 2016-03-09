package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// GenerateConfig generates a starting configuration file to be worked from
func GenerateConfig(file string) {

	config := Config{}
	config.Registry = RegistryConfig{}

	r := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter the URL of the Docker Registry (https://docker.something.io): ")
	config.Registry.URL, _ = r.ReadString('\n')
	fmt.Print("Please enter the username to access the Docker Registry with: ")
	config.Registry.Username, _ = r.ReadString('\n')
	fmt.Print("Please enter the password for the Docker Registry: ")
	password, _ := terminal.ReadPassword(syscall.Stdin)
	config.Registry.Password = string(password)
	fmt.Println()

	// Trim whitespace
	config.Registry.URL = strings.TrimSpace(config.Registry.URL)
	config.Registry.Username = strings.TrimSpace(config.Registry.Username)
	config.Registry.Password = strings.TrimSpace(config.Registry.Password)

	// Quay.io config
	config.Quay = QuayConfig{}
	fmt.Print("Please enter the namespace for the quay.io registry (what comes after quay.io/): ")
	config.Quay.Namespace, _ = r.ReadString('\n')
	fmt.Print("Please enter the username to access the quay.io Registry with: ")
	config.Quay.Username, _ = r.ReadString('\n')
	fmt.Print("Please enter the password for the quay.io Registry: ")
	password, _ = terminal.ReadPassword(syscall.Stdin)
	config.Quay.Password = string(password)
	fmt.Println()

	// Trim whitespace
	config.Quay.Namespace = strings.TrimSpace(config.Quay.Namespace)
	config.Quay.Username = strings.TrimSpace(config.Quay.Username)
	config.Quay.Password = strings.TrimSpace(config.Quay.Password)

	fmt.Println("Populating image information from the Docker Registry......")
	err := PopulateImages(&config)
	if err != nil {
		fmt.Printf("There was an error while populating information: %s\n", err)
		fmt.Println("Writing an incomplete config file....")
	} else {
		fmt.Println("Complete!")
	}

	json, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		panic("Error while writing config to file")
	}

	ioutil.WriteFile(file, json, 0644)
	fmt.Printf("Your configuration file has been generated and saved to %s\n", file)

}

// PopulateImages is used to populate image configuration information
// into the passed in config object based on the given registry credentials
func PopulateImages(c *Config) error {

	c.Images = make([]ImageConfig, 0)

	client := &http.Client{}
	req, err := http.NewRequest("GET", c.Registry.URL+"/v2/_catalog", nil)
	if err != nil {
		return fmt.Errorf("Failed to get the catalog from the Registry. Error: %s", err)
	}

	req.SetBasicAuth(c.Registry.Username, c.Registry.Password)

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Request to the regsitry failed: %s", err)
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	images := catalogResponse{}
	err = d.Decode(&images)
	if err != nil {
		return fmt.Errorf("Unable to parse Registry response: %s", err)
	}

	for _, imageName := range images.Repositories {
		imageConfig := ImageConfig{
			Name:      imageName,
			NewName:   removeNamespace(imageName),
			AdminTeam: "",
			Public:    true,
		}
		c.Images = append(c.Images, imageConfig)
	}

	return nil

}

// removeNamespace is used to generate a non namespaced name for an image
// since quay does not allow additional namespacing. The algorithm is as follows:
//   If the namespace and image are different then replace / with -
//     For example: blah/testing becomes blah-testing
//   If the namespace and image name are the same then remove the namespace
//     For example: blah/blah becomes blah
func removeNamespace(imageName string) string {

	parts := strings.Split(imageName, "/")
	newImageName := ""

	for key, value := range parts {
		if key != len(parts)-1 { // If we aren't on the last item
			if value != parts[key+1] { // If this value is different than the next value
				newImageName = newImageName + value + "-"
			}
		} else {
			newImageName = newImageName + value
		}
	}

	return newImageName
}
