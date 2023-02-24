package certscli

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
)

// CACommands returns the list of CA commands
func CACommands(ctx context.Context, f svcconfig.AppFolders) *cli.Command {

	cmd := &cli.Command{
		// ca view | ca create | ca renew
		Name:      "ca",
		Usage:     "View or create the CA certificate",
		ArgsUsage: ";", //(no args)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "certs",
				Usage:       "Path to certificate `folder`.",
				Value:       f.Certs,
				Destination: &f.Certs,
			},
		},
		Subcommands: []*cli.Command{
			CreateCACommand(ctx, f.Certs),
			ViewCACommand(ctx, f.Certs),
		},
	}
	return cmd
}

// CreateCACommand create the Hub self-signed CA, valid for 5 years
// This does not require any services to run.
// After creating a new CA, services have to be restarted.
//
//	hubcli certs ca [--certs=CertFolder]  [--hostname=hostname]
func CreateCACommand(_ context.Context, certsFolder string) *cli.Command {
	var force = false
	var validityDays = 365 * 5

	return &cli.Command{
		Name:      "createca",
		Aliases:   []string{"crca"},
		Usage:     "Create Hub CA",
		UsageText: "Create a new self-signed Hub CA certificate and private key",
		Category:  "certs",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "days",
				Usage:       "Number of `days` the certificate is valid.",
				Value:       validityDays,
				Destination: &validityDays,
			},
			&cli.BoolFlag{
				Name:        "force",
				Usage:       "Force overwrites an existing certificate and key.",
				Destination: &force,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() > 0 {
				return fmt.Errorf("unexpected argument(s) '%s'", cCtx.Args().First())
			}
			err := HandleCreateCACert(certsFolder, cCtx.Int("days"), cCtx.Bool("force"))
			return err
		},
	}
}

// ViewCACommand shows info on the Hub self-signed CA
// This does not require any services to run.
//
//	hubcli ca [--certs=CertFolder] view
func ViewCACommand(ctx context.Context, certsFolder string) *cli.Command {

	return &cli.Command{
		Name:      "viewca",
		Aliases:   []string{"vca"},
		Category:  "certs",
		Usage:     "View CA certificate info",
		UsageText: "View Hub CA certificate info. Used to create all other certificates.",

		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() > 0 {
				return fmt.Errorf("unexpected argument(s) '%s'", cCtx.Args().First())
			}
			err := HandleViewCACert(ctx, certsFolder)
			return err
		},
	}
}

// HandleCreateCACert generates the hub self-signed CA private key and certificate
// in the given folder.
// Use force to create the folder and overwrite existing certificate if it exists
func HandleCreateCACert(certsFolder string, validityDays int, force bool) error {
	caCertPath := path.Join(certsFolder, hubapi.DefaultCaCertFile)
	caKeyPath := path.Join(certsFolder, hubapi.DefaultCaKeyFile)

	// folder doesn't exist
	if _, err := os.Stat(certsFolder); err != nil {
		if force {
			_ = os.Mkdir(certsFolder, 0744)
		} else {
			return fmt.Errorf("certificate folder '%s' doesn't exist", certsFolder)
		}
	}
	// do not overwrite existing certificate unless force is used
	if !force {
		if _, err := os.Stat(caCertPath); err == nil {
			return fmt.Errorf("CA certificate already exists in '%s'. Use --force to replace", caCertPath)
		}
		if _, err := os.Stat(caKeyPath); err == nil {
			return fmt.Errorf("CA key alread exists in '%s'", caKeyPath)
		}
	}

	caCert, privKey, err := selfsigned.CreateHubCA(validityDays)
	if err != nil {
		return err
	}
	err = certsclient.SaveX509CertToPEM(caCert, caCertPath)
	if err == nil {
		// this sets permissions to 0400 current user readonly
		err = certsclient.SaveKeysToPEM(privKey, caKeyPath)
	}

	logrus.Infof("Generated CA certificate '%s' and key '%s'\n", caCertPath, caKeyPath)
	return err
}

// HandleViewCACert shows CA certificate information
func HandleViewCACert(ctx context.Context, certsFolder string) error {
	caCertPath := path.Join(certsFolder, hubapi.DefaultCaCertFile)
	_ = ctx

	caCert, err := certsclient.LoadX509CertFromPEM(caCertPath)
	if err != nil {
		logrus.Errorf("Unable to load the CA certificate: %s", err)
		return err
	}
	fmt.Println("CA certificate path: ", caCertPath)
	fmt.Println("  IsCA       : ", caCert.IsCA)
	fmt.Println("  Version    : ", caCert.Version)
	fmt.Println("  Valid until: ", caCert.NotAfter.Format(time.RFC1123Z))
	fmt.Println("  Subject    : ", caCert.Subject.String())
	fmt.Println("  Issuer     : ", caCert.Issuer.String())
	fmt.Println("  DNS names  : ", caCert.DNSNames)
	return nil
}
