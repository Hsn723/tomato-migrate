package cmd

import (
	"github.com/cybozu-go/log"
	"github.com/hsn723/tomato-migrate/pkg"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:               "tomato-migrate -i FILE -o FILE",
		Short:             "convert tomato backups",
		Long:              "backup conversion tool for restoring to same model, same firmware tomato routers",
		Args:              cobra.MatchAll(cobra.NoArgs),
		PersistentPreRunE: setup,
		RunE:              runRoot,
	}

	inFile  string
	outFile string
	parser  pkg.IParser
	verbose bool
)

func init() {
	_ = rootCmd.LocalFlags().MarkHidden("logfile")
	_ = rootCmd.LocalFlags().MarkHidden("loglevel")
	_ = rootCmd.LocalFlags().MarkHidden("logformat")
	rootCmd.Flags().StringVarP(&inFile, "in", "i", "", "input file")
	rootCmd.Flags().StringVarP(&outFile, "out", "o", "", "output file")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func setup(_ *cobra.Command, _ []string) error {
	parser = pkg.CfgParser{
		Fs: afero.NewOsFs(),
	}
	logger := log.DefaultLogger()
	if verbose {
		logger.SetThreshold(log.LvInfo)
	} else {
		logger.SetThreshold(log.LvError)
	}
	return nil
}

func runRoot(cmd *cobra.Command, args []string) error {
	_ = log.Info("backing up destination file", map[string]interface{}{
		"file": outFile,
	})
	if err := parser.BackupFile(outFile); err != nil {
		return err
	}

	_ = log.Info("opening destination file", map[string]interface{}{
		"file": outFile,
	})
	data, err := parser.ReadFile(outFile)
	if err != nil {
		return err
	}

	_ = log.Info("retrieving device map from destination", map[string]interface{}{
		"file": outFile,
	})
	destMap := parser.FindDeviceAddresses(string(data))

	_ = log.Info("opening source file", map[string]interface{}{
		"file": inFile,
	})
	source, err := parser.ReadFile(inFile)
	if err != nil {
		return err
	}

	_ = log.Info("retrieving device map from source", map[string]interface{}{
		"file": inFile,
	})
	origMap := parser.FindDeviceAddresses(string(source))

	_ = log.Info("generating MAC address map", nil)
	convertMap := parser.GetMappings(origMap, destMap)

	dest := parser.RemapDevices(string(source), convertMap)
	_ = log.Info("writing remapped file", map[string]interface{}{
		"file": outFile,
	})
	return parser.WriteFile(outFile, dest)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.ErrorExit(err)
	}
}
