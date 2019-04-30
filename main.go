package main

import (
	"github.com/mngharbi/DMPC/daemon"
	dmpcCli "github.com/mngharbi/DMPC/cli"
	"github.com/urfave/cli"
	"log"
	"os"
	"time"
)

func main() {
	app := cli.NewApp()
	app.Name = "DMPC"
	app.Version = "0.0.1"
	app.Compiled = time.Now()
	app.EnableBashCompletion = true
	app.Authors = []cli.Author{
		{
			Name:  "Nizar Gharbi",
			Email: "email@mngharbi.com",
		},
	}
	app.Copyright = "(c) 2018 DMPC"
	app.Usage = "Distributed Multiuser Private Channels"
	app.UsageText = ""

	app.Commands = []cli.Command{
		{
			Name:    "install",
			Aliases: []string{"i"},
			Usage:   "Configure DMPC",
			Action: func(c *cli.Context) error {
				dmpcCli.Install()
				return nil
			},
		},
		{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "Start processing daemon",
			Action: func(c *cli.Context) error {
				daemon.Start()
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
