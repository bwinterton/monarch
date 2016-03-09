package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "Monarch"
	app.Usage = "A tool to migrate Docker Registry images to quay.io"
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
	}

	app.Run(os.Args)

}

func (env *httpClientEnv) getImageTags(imageName string) image {
	req, err := http.NewRequest("GET", env.RegistryURL+"/v2/"+imageName+"/tags/list", nil)
	req.SetBasicAuth(env.Username, env.Password)

	res, err := env.Client.Do(req)
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
