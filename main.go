package main

import (
	"github.com/mngharbi/DMPC/daemon"
	"github.com/mngharbi/DMPC/core"
	dmpcCli "github.com/mngharbi/DMPC/cli"
	"github.com/urfave/cli"
	"log"
	"os"
	"time"
)

func main() {
	app := cli.NewApp()
	app.Name = "DMPC"
	app.Version = core.VersionString
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
			Name:    "operation",
			Aliases: []string{"o"},
			Usage:   "Operation related commands",
			Subcommands: []cli.Command{
				{
					Name:    "sign",
					Usage:   "Sign operation as issuer/certifier",
					Action: func(c *cli.Context) error {
						dmpcCli.ReadAndSignOperation(c.Bool("issue"), c.Bool("certify"))
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
				{
					Name:    "generate",
					Aliases: []string{"g"},
					Usage:   "Generate operations",
					Subcommands: []cli.Command{
						{
							Name:    "user",
							Usage:   "Generate user operations",
							Subcommands: []cli.Command{
								{
									Name:    "create",
									Usage:   "Generate user creation operation from user object",
									Action: func(c *cli.Context) error {
										dmpcCli.GenerateUserCreateOperation()
										return nil
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name:    "user",
			Aliases: []string{"o"},
			Usage:   "User related commands",
			Subcommands: []cli.Command{
				{
					Name:    "generate",
					Aliases: []string{"g"},
					Usage:   "Generate user",
					Action: func(c *cli.Context) error {
						dmpcCli.GenerateUserObject()
						return nil
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
