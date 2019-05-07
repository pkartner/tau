package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	ioa "github.com/iotaledger/iota.go/api"
	"github.com/urfave/cli"

	"github.com/pkartner/tau"
)

const defaultNodeURL = "https://node01.iotatoken.nl:443"
const requestURL = "http://127.0.0.1:8080/"

func createTangleID(c *cli.Context) {
	seed := c.String("seed")
	if seed == "" {
		fmt.Println("A seed is required!")
		return
	}

	nodeURL := defaultNodeURL
	if flagNodeURL := c.String("node"); flagNodeURL != "" {
		nodeURL = flagNodeURL
	}
	api, err := ioa.ComposeAPI(ioa.HTTPClientSettings{
		URI: nodeURL,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(seed)

	tau.CreateOrUpdateTangleID(api, seed, tau.OptionalIDFields{
		Email: "test@test.nl",
	})
}

func request(c *cli.Context) {
	seed := c.String("seed")
	if seed == "" {
		fmt.Println("A seed is required!")
		return
	}

	request, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		panic(err)
	}

	err = tau.SignRequest(seed, request)
	if err != nil {
		panic(err)
	}

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println("The email associated with this account is")
	fmt.Println(string(body))
}

func commands() []cli.Command {
	return []cli.Command{
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "Create a new entry on the tangle using your seed",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "seed, s",
					Usage: "Seed to be used for generating your private/public key and reference address. This can be any string, a longer seed is more secure.",
				},
				cli.StringFlag{
					Name:  "node, n",
					Value: defaultNodeURL,
					Usage: "URL of a iota node for interacting with the tangle.",
				},
			},
			Action: createTangleID,
		},
		{
			Name:    "request",
			Aliases: []string{"r"},
			Usage:   "Make a request to the server example using the seed to authenticate yourself through the tangle",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "seed, s",
					Usage: "Seed you used when you created an entry on the tangle through the create command.",
				},
			},
			Action: request,
		},
	}
}

func main() {
	app := cli.NewApp()
	app.Commands = commands()

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
