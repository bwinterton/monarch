package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

// Simulate simulates what monarch would do if the Migrate command were to
// be run with the given configuration file without actually performing any actions
func Simulate(filename string) {

	c, err := ParseConfig(filename)
	if err != nil {
		fmt.Printf("The configuration file could not be parsed: %s\nExiting now....", err)
	}

	fmt.Println("Simulating Docker Registry -> Quay.io migration....")

	fmt.Printf("Migrating %d images....\n", len(c.Images))

	fmt.Printf("\n------------------------------------------\n\n")

	for _, image := range c.Images {
		fmt.Printf("Migrating image: %s\n", image.Name)
		fmt.Printf("Name: %s/%s -> quay.io/%s/%s\n", c.Registry.URL, image.Name, c.Quay.Namespace, image.NewName)
		fmt.Printf("Would have migrated %d tags\n", len(c.Registry.getImageTags(image.Name).Tags))
		fmt.Printf("Would have set the repository to public: %t\n", image.Public)

		if image.AdminTeam != "" {
			fmt.Printf("Admin team would have been set to: %s\n", image.AdminTeam)
		} else {
			fmt.Println("No admin team would have been set")
		}

		fmt.Printf("\n------------------------------------------\n\n")
	}

	fmt.Println("Simulation complete")

}

// Migrate runs the migration based on the configuration file
func Migrate(filename string, dConfig DockerConfig) {

	c, err := ParseConfig(filename)
	if err != nil {
		fmt.Printf("The configuration file could not be parsed: %s\nExiting now....", err)
	}

	dConfig.Registry = &c.Registry
	dConfig.Quay = &c.Quay
	dConfig.Registry.Base = strings.TrimPrefix(dConfig.Registry.URL, "https://")
	dConfig.Registry.Base = strings.TrimPrefix(dConfig.Registry.Base, "http://") // in case it actually had http://

	fmt.Println("Starting Docker Registry -> Quay.io migration....")

	fmt.Printf("\n------------------------------------------\n\n")

	for key, image := range c.Images {
		fmt.Printf("Migrating image: %s (image %d/%d)\n", image.Name, key+1, len(c.Images))
		iTags := c.Registry.getImageTags(image.Name).Tags
		fmt.Printf("Migrating %d tags\n", len(iTags))

		for tagKey, tag := range iTags {
			fmt.Printf("Migrating tag %d/%d\n", tagKey+1, len(iTags))
			fmt.Printf("Pulling tag: %s\n--------\n\n", tag)
			err = dConfig.pullImage(image.Name, tag)
			if err != nil {
				fmt.Printf("ERROR: Could not pull tag: %s Message: %s\n", tag, err)
				continue
			}

			fmt.Printf("Tagging image %s/%s:%s -> quay.io/%s/%s:%s\n", c.Registry.URL, image.Name, tag, c.Quay.Namespace, image.NewName, tag)
			err = dConfig.tagImage(image.Name, image.NewName, tag)

			fmt.Printf("Pushing tag: %s\n--------\n\n", tag)
			err = dConfig.pushImage(image.NewName, tag)
			if err != nil {
				fmt.Printf("ERROR: Could not push tag: %s Message: %s\n", tag, err)
			}
		}

		if image.Public {
			fmt.Printf("Making %s image public\n", image.NewName)
			err = dConfig.makePublic(image.NewName)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
			}
		}

		if image.AdminTeam != "" {
			fmt.Printf("Assigning %s team to %s image\n", image.AdminTeam, image.NewName)
			err = dConfig.assignTeam(image.NewName, image.AdminTeam)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
			}
		}

		fmt.Printf("Migration of %s image complete!!\n", image.Name)
		fmt.Printf("\n------------------------------------------\n\n")
	}

}

func newDockerConfig(endpoint string, dockerMachine bool) DockerConfig {
	config := DockerConfig{}

	if dockerMachine {
		config.Client, _ = docker.NewClientFromEnv()
	} else {
		config.Client, _ = docker.NewClient(endpoint)
	}

	config.Auth, _ = docker.NewAuthConfigurationsFromDockerCfg()

	return config
}

func (c *DockerConfig) pullImage(imageName, tag string) error {
	pullOptions := docker.PullImageOptions{}

	pullOptions.Registry = c.Registry.Base
	pullOptions.Repository = imageName
	pullOptions.Tag = tag
	pullOptions.OutputStream = os.Stderr

	cmd := exec.Command("docker", "pull", pullOptions.Registry+"/"+imageName+":"+tag)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	//err := c.Client.PullImage(pullOptions, c.Auth.Configs[pullOptions.Registry])
	if err != nil {
		return fmt.Errorf("Error while pulling image: %s\n", err)
	}

	fmt.Println()

	return nil
}

func (c *DockerConfig) tagImage(oldName, newName, tag string) error {

	cmd := exec.Command("docker", "tag", c.Registry.Base+"/"+oldName+":"+tag, "quay.io/"+c.Quay.Namespace+"/"+newName+":"+tag)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error while retagging image: %s\n", err)
	}

	return nil
}

func (c *DockerConfig) pushImage(imageName, tag string) error {

	cmd := exec.Command("docker", "push", "quay.io/"+c.Quay.Namespace+"/"+imageName+":"+tag)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error while pushing image: %s\n", err)
	}

	fmt.Println()

	return nil
}

func (c *DockerConfig) makePublic(imageName string) error {

	reqBody := []byte(`{"visibility":"public"}`)
	url := "https://quay.io/api/v1/repository/" + c.Quay.Namespace + "/" + imageName + "/changevisibility"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))

	req.Header.Set("authorization", "Bearer "+c.Quay.OAuthToken)
	req.Header.Set("content-type", "application/json")

	res, err := new(http.Client).Do(req)
	if err != nil {
		return fmt.Errorf("Error while making image public: Error:%s\n", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Error while making image public: Code: %d", res.StatusCode)
	}

	return nil
}

func (c *DockerConfig) assignTeam(imageName, team string) error {

	reqBody := []byte(`{"role":"admin"}`)
	url := "https://quay.io/api/v1/repository/" + c.Quay.Namespace + "/" + imageName + "/permissions/team/" + team
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqBody))

	req.Header.Set("authorization", "Bearer "+c.Quay.OAuthToken)
	req.Header.Set("content-type", "application/json")

	res, err := new(http.Client).Do(req)
	if err != nil {
		return fmt.Errorf("Error while assigning team to image: Error:%s\n", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Error while assigning team to image: Code: %d", res.StatusCode)
	}

	return nil

}

func (c *RegistryConfig) getImageTags(imageName string) image {
	req, err := http.NewRequest("GET", c.URL+"/v2/"+imageName+"/tags/list", nil)
	req.SetBasicAuth(c.Username, c.Password)

	res, err := new(http.Client).Do(req)
	if err != nil {
		panic("Unable to get tag list for image: " + imageName)
	}
	defer res.Body.Close()

	image := image{}
	d := json.NewDecoder(res.Body)
	err = d.Decode(&image)
	if err != nil {
		panic("Unable to parse tag list for image: " + imageName)
	}

	return image

}
