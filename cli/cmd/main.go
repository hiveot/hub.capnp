package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/wostzone/hub/launcher/cmd/certcli"
	"github.com/wostzone/hub/svc/certsvc/certconfig"
)

const Version = `0.4-alpha`

var binFolder string
var homeFolder string
var configFolder string
var certFolder string

// CLI Main entry
func main() {
	logrus.SetLevel(logrus.InfoLevel)
	binFolder = path.Dir(os.Args[0])
	homeFolder = path.Dir(binFolder)
	configFolder = path.Join(homeFolder, "config")
	certFolder = path.Join(homeFolder, "certs")

	app := &cli.App{
		EnableBashCompletion: true,
		Name:                 "hubcli",
		Usage:                "Hub Commandline Interface",
		Version:              Version,
		Commands: []*cli.Command{
			// certs ca | client | device --certs=folder --pubkey=path ID
			{
				Name:  "certs",
				Usage: "Create certificates",
				Subcommands: []*cli.Command{
					GetCertsCreateCACommand(),
					GetCertsCreateServiceCommand(),
					GetCertsCreateClientCommand(),
					GetCertsCreateDeviceCommand(),
				},
			},
			//hub prov add|list  <deviceID> <secret>
			{
				Name:  "prov",
				Usage: "IoT device provisioning",
				Subcommands: []*cli.Command{
					GetProvAddCommand(),
				},
			},
			// service management
			{
				Name:  "svc",
				Usage: "Services management",
				Subcommands: []*cli.Command{
					GetSvcListCommand(),
					GetSvcStartCommand(),
					GetSvcStopCommand(),
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// GetCertsCreateCACommand
// hubcli certs ca [--certs=CertFolder]  [--hostname=hostname]
func GetCertsCreateCACommand() *cli.Command {
	var hostname = "localhost"
	var force = false
	var validityDays = certconfig.DefaultCACertDurationDays
	return &cli.Command{
		Name:      "ca",
		Usage:     "Create Hub CA certificate and key",
		ArgsUsage: "(no args)",
		//Category: "create",
		//ArgsUsage: "--certs=folder --hostname=name  --force",
		//UsageText: "--certs=folder --hostname=name  --force",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "certs",
				Usage:       "Path to certificate `folder`.",
				Value:       certFolder,
				Destination: &certFolder,
			},
			&cli.StringFlag{
				Name:        "hostname",
				Usage:       "host `name` or IP the certificate is valid for.",
				Value:       hostname,
				Destination: &hostname,
			},
			&cli.IntFlag{
				Name:        "days",
				Usage:       "Number of `days` the certificate is valid.",
				Value:       validityDays,
				Destination: &validityDays,
			},
			&cli.BoolFlag{
				Name:        "force",
				Usage:       "Force overwrites an existing certificate and key.",
				Aliases:     []string{"f"},
				Destination: &force,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() > 0 {
				return fmt.Errorf("unexpected argument(s) '%s'", cCtx.Args().First())
			}
			err := certcli.HandleCreateCACert(
				cCtx.String("certs"),
				cCtx.String("hostname"),
				cCtx.Int("days"),
				cCtx.Bool("force"),
			)
			//logrus.Infof("CreatingCA certificate in '%s' for host '%s'",
			//cCtx.String("certs"), cCtx.String("hostname"))
			return err
		},
	}
}

// GetCertsCreateClientCommand
// hubcli certs client [--certs=CertFolder --pubkey=pubkeyfile] <loginID>
func GetCertsCreateClientCommand() *cli.Command {
	validityDays := certconfig.DefaultClientCertDurationDays

	return &cli.Command{
		Name:      "client",
		Usage:     "Create a client certificate",
		ArgsUsage: "<loginID>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "certs",
				Usage:       "Path to certificate `folder`.",
				Value:       certFolder,
				Destination: &certFolder,
			},
			&cli.StringFlag{
				Name:  "pubkey",
				Usage: "`file` with client public or private key in PEM format. When omitted a public/private key pair will be generated.",
			},
			&cli.IntFlag{
				Name:        "days",
				Usage:       "Number of days the certificate is valid.",
				Value:       validityDays,
				Destination: &validityDays,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() == 0 {
				return fmt.Errorf("Missing client login ID")
			}
			loginID := cCtx.Args().Get(0)
			pubKeyFile := cCtx.String("pubkey")
			err := certcli.HandleCreateClientCert(certFolder, loginID, pubKeyFile, validityDays)
			return err
		},
	}
}

// GetCertsCreateDeviceCommand
// hubcli certs device [--certs=CertFolder] --pubkey=pubkeyfile <deviceID>
func GetCertsCreateDeviceCommand() *cli.Command {
	validityDays := certconfig.DefaultDeviceCertDurationDays

	return &cli.Command{
		Name:      "device",
		Usage:     "Create an IoT device certificate",
		ArgsUsage: "<deviceID>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "certs",
				Usage:       "Path to certificates `folder`.",
				Value:       certFolder,
				Destination: &certFolder,
			},
			&cli.StringFlag{
				Name:  "pubkey",
				Usage: "`file` with device public or private key in PEM format. When omitted a public/private key pair will be generated.",
			},
			&cli.IntFlag{
				Name:        "days",
				Usage:       "Number of days the certificate is valid.",
				Value:       validityDays,
				Destination: &validityDays,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() == 0 {
				return fmt.Errorf("Missing device ID")
			}
			deviceID := cCtx.Args().Get(0)
			pubKeyFile := cCtx.String("pubkey")
			err := certcli.HandleCreateDeviceCert(certFolder, deviceID, pubKeyFile, validityDays)
			return err
		},
	}
}

// GetCertsCreateServiceCommand
// hubcli certs service [--certs=CertFolder --pubkey=pubkeyfile] <serviceID>
func GetCertsCreateServiceCommand() *cli.Command {
	validityDays := certconfig.DefaultServiceCertDurationDays
	ipAddr := ""

	return &cli.Command{
		Name:      "service",
		Usage:     "Create a service certificate",
		ArgsUsage: "<serviceID>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "certs",
				Usage:       "Path to certificate `folder` containing CA certificate and keys.",
				Value:       certFolder,
				Destination: &certFolder,
			},
			&cli.StringFlag{
				Name:  "pubkey",
				Usage: "`file` with service public or private key in PEM format. When omitted a public/private key pair will be generated.",
			},
			&cli.StringFlag{
				Name:        "ipAddr",
				Usage:       "Optional service IP address in addition to localhost.",
				Destination: &ipAddr,
			},
			&cli.IntFlag{
				Name:        "days",
				Usage:       "Number of days the certificate is valid.",
				Value:       validityDays,
				Destination: &validityDays,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() == 0 {
				return fmt.Errorf("Missing service ID")
			}
			serviceID := cCtx.Args().Get(0)
			pubKeyFile := cCtx.String("pubkey")
			err := certcli.HandleCreateServiceCert(certFolder, serviceID, ipAddr, pubKeyFile, validityDays)
			return err
		},
	}
}

// GetProvAddCommand
// prov add [--secrets=folder] <deviceID> <oobsecret>
func GetProvAddCommand() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add an out-of-band device provisioning secret for automatic provisioning",
		ArgsUsage: "<deviceID> <oobSecret>",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 2 {
				return fmt.Errorf("expected 2 arguments. Got %d instead", cCtx.NArg())
			}
			fmt.Println("Adding secret for device: ", cCtx.Args().First())
			return nil
		},
	}
}

// GetSvcListCommand
// svc list
func GetSvcListCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List Hub services and their status",
		ArgsUsage: ".",
		Action: func(cCtx *cli.Context) error {
			fmt.Println("Listing services")
			return nil
		},
	}
}

// GetSvcStartCommand
// svc start
func GetSvcStartCommand() *cli.Command {
	return &cli.Command{
		Name:      "start",
		Usage:     "Start a Hub service",
		ArgsUsage: "<serviceName>",
		Action: func(cCtx *cli.Context) error {
			fmt.Println("Starting service: ", cCtx.Args().First())
			return nil
		},
	}
}

// GetSvcStopCommand
// svc stop
func GetSvcStopCommand() *cli.Command {
	return &cli.Command{
		Name:      "stop",
		Usage:     "Stop a Hub service",
		ArgsUsage: "<serviceName>",
		Action: func(cCtx *cli.Context) error {
			fmt.Println("Stopping service: ", cCtx.Args().First())
			return nil
		},
	}
}
