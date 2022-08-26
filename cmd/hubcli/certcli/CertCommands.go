// Package certcli with certificate command handling
package certcli

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/wostzone/hub/internal/folders"

	"github.com/wostzone/hub/pkg/svc/certsvc/certconfig"
)

// GetCertCommands returns the certificate handling commands
func GetCertCommands(homeFolder string) *cli.Command {
	certFolder := folders.GetFolders(homeFolder).Certs

	cmd := &cli.Command{
		// certs ca | client | device --certs=folder --pubkey=path ID

		Name:  "cert",
		Usage: "Create certificates",
		Subcommands: []*cli.Command{
			GetCertCreateCACommand(certFolder),
			GetCertCreateServiceCommand(certFolder),
			GetCertCreateClientCommand(certFolder),
			GetCertCreateDeviceCommand(certFolder),
			GetCertShowInfoCommand(),
		},
	}
	return cmd
}

// GetCertCreateCACommand
// hubcli certs ca [--certs=CertFolder]  [--hostname=hostname]
func GetCertCreateCACommand(certFolder string) *cli.Command {
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
			err := HandleCreateCACert(
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

// GetCertCreateClientCommand
// hubcli certs client [--certs=CertFolder --pubkey=pubkeyfile] <loginID>
func GetCertCreateClientCommand(certFolder string) *cli.Command {
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
			err := HandleCreateClientCert(certFolder, loginID, pubKeyFile, validityDays)
			return err
		},
	}
}

// GetCertCreateDeviceCommand
// hubcli certs device [--certs=CertFolder] --pubkey=pubkeyfile <deviceID>
func GetCertCreateDeviceCommand(certFolder string) *cli.Command {
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
			err := HandleCreateDeviceCert(certFolder, deviceID, pubKeyFile, validityDays)
			return err
		},
	}
}

// GetCertCreateServiceCommand
// hubcli certs service [--certs=CertFolder --pubkey=pubkeyfile] <serviceID>
func GetCertCreateServiceCommand(certFolder string) *cli.Command {
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
			err := HandleCreateServiceCert(certFolder, serviceID, ipAddr, pubKeyFile, validityDays)
			return err
		},
	}
}
func GetCertShowInfoCommand() *cli.Command {
	return &cli.Command{
		Name:      "info",
		Usage:     "Show certificate info",
		ArgsUsage: "<certFile>",
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() != 1 {
				return fmt.Errorf("expected 1 argument. Got %d instead", cCtx.NArg())
			}
			HandleShowCertInfo(cCtx.Args().First())
			return nil
		},
	}
}
