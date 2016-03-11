package main

import "github.com/fsouza/go-dockerclient"

// Config struct to import the config file
type Config struct {
	Quay     QuayConfig
	Registry RegistryConfig
	Images   []ImageConfig
}

// ImageConfig is the Image struct found in the config file
type ImageConfig struct {
	Name      string
	NewName   string
	AdminTeam string
	Public    bool
}

// RegistryConfig holds the configuration for the private docker registry
type RegistryConfig struct {
	URL      string
	Username string
	Password string
	Base     string
}

// QuayConfig holds the configuration for the quay repository to move to
type QuayConfig struct {
	Namespace  string // The quay.io namespace to push the images to
	OAuthToken string
}

// DockerConfig contains the configuration for the docker client
type DockerConfig struct {
	Client   *docker.Client
	Auth     *docker.AuthConfigurations
	Registry *RegistryConfig
	Quay     *QuayConfig
}
