package main

import (
	"net/http"
	"os"

	"github.com/codegangsta/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "Monarch"
	app.Usage = "A tool to migrate Docker Registry images to quay.io"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "config.json",
			Usage: "The location of the config file to use. Optional, defaults to config.json",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "generateConfig",
			Aliases: []string{"gc"},
			Usage:   "Generates a configuration file to work from",
			Action: func(c *cli.Context) {
				GenerateConfig(c.String("filename"))
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "filename, f",
					Value: "config.json",
					Usage: "The file to write the config to. Optional, defaults to config.json",
				},
			},
		},
		{
			Name:    "validateConfig",
			Aliases: []string{"vc", "validate"},
			Action: func(c *cli.Context) {
				ValidateFile(c.Parent().String("config"))
			},
		},
		{
			Name:    "simulate",
			Aliases: []string{"sim", "s"},
			Action: func(c *cli.Context) {
				Simulate(c.Parent().String("config"))
			},
		},
		{
			Name:    "migrate",
			Aliases: []string{"mig", "s"},
			Action: func(c *cli.Context) {
				dConfig := newDockerConfig(c.String("docker-endpoint"), c.Bool("docker-machine"))
				Migrate(c.Parent().String("config"), dConfig)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "docker-endpoint",
					Value: "unix:///var/run/docker.sock",
					Usage: "The docker endpoint to use for the docker client. Optional, defaults to unix://var/run/docker.sock",
				},
				cli.BoolFlag{
					Name:  "docker-machine",
					Usage: "Use this flag if you are running docker machine and want to use the docker-machine client",
				},
			},
		},
	}

	app.Run(os.Args)

}

type catalogResponse struct {
	Repositories []string
}

type image struct {
	Name string
	Tags []string
}

type catalog struct {
	Images []image
}

type httpClientEnv struct {
	Client      *http.Client
	RegistryURL string
	Username    string
	Password    string
}
