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
	includeAllVariants       bool
	saveDownloads            bool
)

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "output-file", "o", "clinvar_assessments.csv", "Output file to write")
	rootCmd.Flags().StringVarP(&clinvarVcfFile, "clinvar-vcf", "c", LatestClinvarVCFUrl, "ClinVar vcf file, leave blank to download latest")
	rootCmd.Flags().BoolVarP(&includeAllVariants, "include-all", "a", false, "Include low quality, non passing variants. Will use PASSing variants by default")
	rootCmd.Flags().BoolVarP(&saveDownloads, "keep-downloads", "k", false, "Keep the ClinVar downloaded files when complete, will be deleted by default")
	rootCmd.Flags().StringVarP(&clinvarSubmissionSummary, "clinvar-submissions", "s", LatestClinvarSubmissionSummaryUrl, "ClinVar submission summary file, leave blank to download latest")
}

var rootCmd = &cobra.Command{
	Use:           "clinvar-matcher [vcfFile]",
	Short:         "clinvar-matcher is a tool to match your vcf with the latest ClinVar",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reportConfig := matcher.ReportConfig{
			SourceVcfPath:         args[0],
			ClinvarVcfPath:        clinvarVcfFile,
			ClinvarSubmissionPath: clinvarSubmissionSummary,
			IncludeAllVariants:    includeAllVariants,
			OutputFile:            outputFile,
			SaveDownloads:         saveDownloads,
		}
		return matcher.GenerateAssessmentReport(reportConfig)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
