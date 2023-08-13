package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/dsh2dsh/fio/internal/app"
)

const yamlFileName = ".fio.yaml"

var (
	cfg     *app.Config
	cfgFile string

	fromDate string
	toDate   string
	oneMonth string

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "fio [input.csv]",
		Short: "Generator of a report of money expenses using CSV from Fio banka.",
		Long: `This program reads a CSV file or stdin, processes it and outputs an aggreagated
report of your expenses.`,
		Run: func(cmd *cobra.Command, args []string) {
			inputFile := os.Stdin
			if len(args) > 0 {
				file, err := os.Open(args[0])
				cobra.CheckErr(err)
				defer file.Close()
				inputFile = file
			}

			report := withFromToDates(app.NewReport(cfg))
			cobra.CheckErr(report.Parse(inputFile))
			if report.Data().Count() > 0 {
				cobra.CheckErr(report.Print())
			} else {
				fmt.Fprintln(os.Stderr, "Nothing found for given dates.")
			}
		},
	}
)

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main(). It only needs to happen once to
// the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		fmt.Sprintf("config file (default is %s)", yamlFileName))

	rootCmd.PersistentFlags().StringVar(&fromDate, "from-date", "",
		"skip payments before given date (in format YYYY-MM-DD)")
	rootCmd.PersistentFlags().StringVar(&toDate, "to-date", "",
		"skip payments after given date (in format YYYY-MM-DD)")

	rootCmd.PersistentFlags().StringVarP(&oneMonth, "month", "m", "",
		"include payments for given month (in format YYYY-MM)")
	rootCmd.MarkFlagsMutuallyExclusive("month", "from-date")
	rootCmd.MarkFlagsMutuallyExclusive("month", "to-date")
}

func initConfig() {
	if cfgFile != "" {
		tryConfig(cfgFile, true)
		return
	}

	tryFiles := []func() (string, error){
		func() (string, error) { // in cur dir
			return yamlFileName, nil
		},
		homeCfgPath,
	}

	for _, nextFileName := range tryFiles {
		path, err := nextFileName()
		if err != nil {
			cobra.CheckErr(err)
		}
		if tryConfig(path, false) {
			return
		}
	}

	cobra.CheckErr(fmt.Errorf("%q not found", yamlFileName))
}

func tryConfig(path string, mustExists bool) bool {
	c, err := app.LoadConfig(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && !mustExists {
			return false
		}
		cobra.CheckErr(err)
	}
	cfg = c
	return true
}

func homeCfgPath() (string, error) {
	if home, err := os.UserHomeDir(); err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	} else {
		return filepath.Join(home, yamlFileName), nil
	}
}

func withFromToDates(report *app.Report) *app.Report {
	if report, ok := withOneMonth(report); ok {
		return report
	}

	if fromDate != "" {
		if d, err := time.Parse("2006-01-02", fromDate); err != nil {
			cobra.CheckErr(err)
		} else {
			report = report.WithFromDate(d)
		}
	}

	if toDate != "" {
		if d, err := time.Parse("2006-01-02", toDate); err != nil {
			cobra.CheckErr(err)
		} else {
			report = report.WithToDate(d)
		}
	}

	return report
}

func withOneMonth(report *app.Report) (*app.Report, bool) {
	if oneMonth == "" {
		return report, false
	}

	d, err := time.Parse("2006-01", oneMonth)
	if err != nil {
		cobra.CheckErr(err)
	}
	report = report.WithFromDate(d).WithToDate(d.AddDate(0, 1, -1))

	return report, true
}
