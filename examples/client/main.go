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

const nodeURL = "https://node01.iotatoken.nl:443"
const requestURL = "http://127.0.0.1:8080/"

func createTangleID(c *cli.Context) {
	seed := c.String("seed")
	if seed == "" {
		fmt.Println("A seed is required!")
		return
	}

	api, err := ioa.ComposeAPI(ioa.HTTPClientSettings{
		URI: nodeURL,
	})
	if err != nil {
		fmt.Println("We got the following error when we tried composing the iota api ", err)
		return
	}

	fmt.Printf("Creating an account on the tangle using \"%s\" as a seed...\n", seed)

	reference, err := tau.CreateOrUpdateTangleID(api, seed, tau.OptionalIDFields{
		// If the user gave a name we extract it and attach it to the tangleID
		Name: c.String("name"),
	})
	if err == tau.ErrCallingTangle {
		// Tangle node is not available right now normally we should have a backup node to try again
		fmt.Println("Something wen't wrong while calling the tangle. You could try to change the nodeURL constant with a healthy node found here https://www.iotatoken.nl/")
		return
	}
	if err != nil {
		// Should not happen so we panic this unhappy flow
		panic(err)
	}
	fmt.Printf("Account succesfully created your reference is %s\n", reference)
	fmt.Printf("Go to https://thetangle.org/address/%s to check it out.", reference)
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
		// If we get a error that means something none input depenend went wrong so we panic
		panic(err)
	}

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		fmt.Println("We could not make a request something wen't wrong on the server or there is something wrong with your connection.")
		return
	}
	if response.StatusCode == http.StatusUnauthorized {
		fmt.Println("You could not be authorized, did you create an account on the tangle?")
		return
	}
	if response.StatusCode == http.StatusServiceUnavailable {
		fmt.Println("Server could currently not authenticate you because it was unavailable. It probably is because the tangle node was not availabe. You could try to change the nodeURL constant with a healthy node found here https://www.iotatoken.nl/ ")
		return
	}
	if response.StatusCode != http.StatusOK {
		fmt.Println("Something went wrong on the server.")
		return
	}

	defer response.Body.Close()
	fmt.Println("You are authorized")
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	name := string(body)
	if name == "" {
		return
	}
	fmt.Printf("The name associated with this id is %s\n", name)
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
					Name:  "name, n",
					Usage: "Name that will be attached to your id",
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
