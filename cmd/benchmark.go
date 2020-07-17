package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	engine "github.com/mrfleap/custom-compression/engine"
	"github.com/spf13/cobra"
)

func init() {
	var maxSearchBuffer int

	var benchmarkCmd = &cobra.Command{
		Use:   "benchmark",
		Short: "Benchmark a file using custom-compressor",
		Long:  `Benchmark a file using custom-compressor`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("please provide 2 parameters: engine and file")
			}
			userEngine := args[0]
			foundEngine := false
			for _, engine := range engine.Engines {
				if engine == userEngine { foundEngine = true }
			}
			if !foundEngine {
				return fmt.Errorf("\"%s\" is not a valid engine type, please choose one of the following:\n\t %s", userEngine, strings.Join(engine.Engines[:], ", "))
			}
			file := args[1]
			if _, err := os.Stat(file); os.IsNotExist(err) {
				return fmt.Errorf("Could not open file (likely does not exist): %s", file)
			} else {
				return nil
			}
		},
		Run: func (cmd *cobra.Command, args []string) {
			userEngine := args[0]
			file := args[1] // Args[0] = file as a string
			engine.BenchmarkFile(userEngine, file, maxSearchBuffer)
		},
	}

	benchmarkCmd.PersistentFlags().IntVarP(&maxSearchBuffer, "max search buffer length", "m", 1024, "maximum length of search buffer")
	rootCmd.AddCommand(benchmarkCmd)
}


