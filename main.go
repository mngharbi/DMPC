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
		{
			Name:    "sign",
			Usage:   "Sign operation as issuer/certifier",
			Action: func(c *cli.Context) error {
				dmpcCli.SignOperation(c.Bool("issue"), c.Bool("certify"))
				return nil
			},   
			Flags: []cli.Flag{
		    	cli.BoolFlag{
		    		Name: "issue, i",
		    		Usage: "Sign as issuer",
		    	},
		    	cli.BoolFlag{
		    		Name: "certify, c",
		    		Usage: "Sign as certifier",
		    	},
		    },
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
