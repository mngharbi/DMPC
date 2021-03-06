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

var (
	channelFlagsMap map[string]cli.Flag = map[string]cli.Flag{
		"channel": cli.StringFlag{
			Name: "channel, c",
			Usage: "Channel id",
		},
		"noencrypt": cli.BoolFlag{
			Name: "noencrypt, ne",
			Usage: "No encryption",
		},
		"nosign": cli.BoolFlag{
			Name: "nosign, ns",
			Usage: "No signature",
		},
		"encrypt": cli.BoolFlag{
			Name: "encrypt, e",
			Usage: "Encrypt operation",
		},
		"sign": cli.BoolFlag{
			Name: "sign, s",
			Usage: "Sign operation",
		},
	}
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
			Name:    "transaction",
			Aliases: []string{"o"},
			Usage:   "Transaction related commands",
			Subcommands: []cli.Command{
				{
					Name:    "generate",
					Aliases: []string{"g"},
					Usage:   "Generate transaction",
					Action: func(c *cli.Context) error {
						dmpcCli.GenerateTransaction(c.Bool("ignoreresult"), c.Bool("statusupdate"), c.Bool("keepalive"), c.StringSlice("recepient"))
						return nil
					},
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name: "ignoreresult, i",
							Usage: "Ignore transaction result",
						},
						cli.BoolFlag{
							Name: "statusupdate, u",
							Usage: "Get transaction status updates",
						},
						cli.BoolFlag{
							Name: "keepalive, k",
							Usage: "Keep connection open after transaction",
						},
						cli.StringSliceFlag{
							Name: "recepient, r",
							Usage: "Encrypt transaction for recepient",
						},
					},
				},
				{
					Name:    "run",
					Aliases: []string{"r"},
					Usage:   "Run transaction",
					Action: func(c *cli.Context) error {
						dmpcCli.ReadAndRunOneTransaction()
						return nil
					},
				},
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
						{
							Name:    "channel",
							Usage:   "Generate channel operations",
							Subcommands: []cli.Command{
								{
									Name:    "read",
									Usage:   "Generate channel read operation",
									Flags: []cli.Flag{
										channelFlagsMap["channel"],
										channelFlagsMap["sign"],
									},
									Action: func(c *cli.Context) error {
										dmpcCli.GenerateChannelReadOperation(c.String("channel"), c.Bool("sign"), c.Bool("sign"))
										return nil
									},
								},
								{
									Name:    "open",
									Usage:   "Generate channel open operation from channel object",
									Flags: []cli.Flag{
										channelFlagsMap["sign"],
									},
									Action: func(c *cli.Context) error {
										dmpcCli.GenerateChannelOpenOperation(c.Bool("sign"), c.Bool("sign"))
										return nil
									},
								},
								{
									Name:    "close",
									Usage:   "Generate channel close operation",
									Flags: []cli.Flag{
										channelFlagsMap["channel"],
										channelFlagsMap["sign"],
										channelFlagsMap["encrypt"],
									},
									Action: func(c *cli.Context) error {
										dmpcCli.GenerateChannelCloseOperation(c.String("channel"), c.Bool("sign"), c.Bool("sign"), c.Bool("encrypt"))
										return nil
									},
								},
								{
									Name:    "listen",
									Usage:   "Generate channel listen operation",
									Flags: []cli.Flag{
										channelFlagsMap["channel"],
										channelFlagsMap["sign"],
									},
									Action: func(c *cli.Context) error {
										dmpcCli.GenerateChannelSubscribeOperation(c.String("channel"), c.Bool("sign"), c.Bool("sign"))
										return nil
									},
								},
								{
									Name:    "message",
									Usage:   "Generate channel message operation",
									Flags: []cli.Flag{
										channelFlagsMap["channel"],
										channelFlagsMap["nosign"],
										channelFlagsMap["noencrypt"],
									},
									Action: func(c *cli.Context) error {
										dmpcCli.GenerateChannelAddMessageOperation(c.String("channel"), !c.Bool("nosign"), !c.Bool("nosign"), !c.Bool("noencrypt"))
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
		{
			Name:    "channel",
			Aliases: []string{"c"},
			Usage:   "Channel related commands",
			Subcommands: []cli.Command{
				{
					Name:    "generate",
					Aliases: []string{"g"},
					Usage:   "Generate channel object",
					Action: func(c *cli.Context) error {
						dmpcCli.GenerateChannelObject()
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
