package cmd

import (
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run testfile",
	Short: "Run the given test",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		test, err := LoadTestSuite(args[0])
		if err != nil {
			log.Fatalf("Error loading test: %s", err)
		}
		log.Infof("running test %s", test.Name)
		results, err := test.RunTestSuite(test)
		if len(results.SuccessTemplates) == len(results.Suite.NucleiTemplates) {
			log.Infof("Test %s passed", test.Name)
		} else if len(results.ErroredTemplates) == 0 {
			log.Warningf("Test %s had %d failures", test.Name, len(results.FailedTemplates))
		} else {
			log.Errorf("Test %s had %d failures and %d errors", test.Name, len(results.FailedTemplates), len(results.ErroredTemplates))
		}
		if len(results.SuccessTemplates) > 0 {
			log.Infof("Successful tests: %v", results.SuccessTemplates)
		}
		if len(results.FailedTemplates) > 0 {
			log.Warningf("Failed tests: %v", results.FailedTemplates)
		}
		if len(results.ErroredTemplates) > 0 {
			log.Errorf("Errored tests: %v", results.ErroredTemplates)
		}
		if err != nil {
			log.Errorf("Error running test %s: %s", test.Name, err)
		}
	},
}
