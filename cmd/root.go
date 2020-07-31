package cmd

import (
	"fmt"
	"os"

	"github.com/kazmiekr/clinvar-matcher/matcher"
	"github.com/spf13/cobra"
)

const (
	LatestClinvarVCFUrl               = "https://ftp.ncbi.nlm.nih.gov/pub/clinvar/vcf_GRCh37/clinvar.vcf.gz"
	LatestClinvarSubmissionSummaryUrl = "https://ftp.ncbi.nlm.nih.gov/pub/clinvar/tab_delimited/submission_summary.txt.gz"
)

var (
	clinvarVcfFile           string
	clinvarSubmissionSummary string
	outputFile               string
)

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "output-file", "o", "clinvar_assessments.csv", "Output file to write")
	rootCmd.Flags().StringVarP(&clinvarVcfFile, "clinvar-vcf", "c", LatestClinvarVCFUrl, "Clinvar vcf file, leave blank to download latest")
	rootCmd.Flags().StringVarP(&clinvarSubmissionSummary, "clinvar-submissions", "s", LatestClinvarSubmissionSummaryUrl, "Clinvar submission summary file, leave blank to download latest")
}

var rootCmd = &cobra.Command{
	Use:           "clinvar-matcher [vcfFile]",
	Short:         "Clinvar Matcher is a tool to match your vcf with the latest Clinvar",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return matcher.GenerateAssessmentReport(args[0], clinvarVcfFile, clinvarSubmissionSummary, outputFile)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
