package matcher

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/kazmiekr/clinvar-matcher/clinvar"
	"github.com/kazmiekr/clinvar-matcher/downloader"
	"github.com/kazmiekr/clinvar-matcher/vcf"

	log "github.com/sirupsen/logrus"
)

const (
	SnpediaLinkPattern = "https://www.snpedia.com/index.php/%s"
	DbSNPLinkPattern   = "https://www.ncbi.nlm.nih.gov/snp/%s"
	ClinvarLinkPattern = "https://www.ncbi.nlm.nih.gov/clinvar/variation/%s/"
)

type ReportConfig struct {
	SourceVcfPath         string
	OutputFile            string
	ClinvarVcfPath        string
	ClinvarSubmissionPath string
	IncludeAllVariants    bool
	SaveDownloads         bool
}

// Will return the clinvar rsid first, otherwise it'll try to use the vcf id
func getRsid(vcfFileId, clinvarRsid string) string {
	if clinvarRsid != "" {
		return fmt.Sprintf("rs%s", clinvarRsid)
	}
	return vcfFileId
}

func isPassingVariantFilter(filter string) bool {
	return strings.ToLower(filter) == "pass"
}

func GenerateAssessmentReport(config ReportConfig) error {

	downloads := make([]string, 0)
	// If user did not specify clinvar VCF file, download latest
	clinvarFile := config.ClinvarVcfPath
	if strings.Index(clinvarFile, "https://") != -1 {
		description := "Downloading Clinvar VCF"
		localFile := path.Base(clinvarFile)
		err := downloader.DownloadFile(localFile, clinvarFile, description)
		if err != nil {
			return err
		}
		downloads = append(downloads, localFile)
		clinvarFile = localFile
	}

	// If user did not specify clinvar submissions file, download latest
	clinvarSubmissionFile := config.ClinvarSubmissionPath
	if strings.Index(clinvarSubmissionFile, "https://") != -1 {
		description := "Downloading Clinvar Submissions"
		localFile := path.Base(clinvarSubmissionFile)
		err := downloader.DownloadFile(localFile, clinvarSubmissionFile, description)
		if err != nil {
			return err
		}
		downloads = append(downloads, localFile)
		clinvarSubmissionFile = localFile
	}

	err := WriteAssessedVariants(config, clinvarFile, clinvarSubmissionFile)
	if err != nil {
		return err
	}

	if config.SaveDownloads == false {
		for _, f := range downloads {
			err = os.Remove(f)
			if err != nil {
				return err
			}
			log.Infof("Deleted downloaded file %s\n", f)
		}
	}
	return err
}

func WriteAssessedVariants(config ReportConfig, localClinvarVcfPath string, localSubmissionPath string) error {
	log.Infof("Loading vcf from %s", config.SourceVcfPath)
	variants, err := vcf.ReadVcf(config.SourceVcfPath)
	if err != nil {
		return err
	}
	log.Infof("Variant Count: %d\n", len(variants))

	clinvarClient, err := clinvar.NewClinvar(localClinvarVcfPath, localSubmissionPath)
	if err != nil {
		return err
	}

	matches := 0
	counts := make(map[string]int)

	resultFile, err := os.Create(config.OutputFile)
	if err != nil {
		return err
	}
	defer resultFile.Close()
	writer := csv.NewWriter(resultFile)
	defer writer.Flush()

	header := []string{
		"Chromosome",
		"Begin",
		"End",
		"Var Type",
		"Quality",
		"Filter",
		"Ref",
		"Alt",
		"Rsid",
		"Zygosity",
		"Clinvar ID",
		"Assessment Count",
		"Max Pathogenicity",
		"# Benign",
		"# Likely Benign",
		"# VUS",
		"# Likely Path",
		"# Pathogenic",
		"# Other",
		"Diseases",
		"Genes",
		"Clinvar Link",
		"dbSNP Link",
		"Snpedia Link",
		"AF_ESP",
		"AF_EXAC",
		"AF_TGP",
	}
	writer.Write(header)

	if config.IncludeAllVariants {
		log.Infof("Including ALL variants, regardless of variant quality")
	} else {
		log.Infof("Filtering only PASSing variants based on VCF Filter")
	}

	for _, line := range variants {
		// Quality filter
		if config.IncludeAllVariants == false && isPassingVariantFilter(line.Filter) == false {
			continue
		}
		if clinvarMatch, ok := clinvarClient.Lookup(clinvar.ToClinvarKey(line)); ok {
			matches++
			pathogenicity := clinvarMatch.Variant.Info["CLNSIG"]
			if _, ok := counts[pathogenicity]; !ok {
				counts[pathogenicity] = 0
			}
			counts[pathogenicity]++

			sigs := make([]string, 0)
			diseases := make([]string, 0)
			genes := make([]string, 0)
			for _, assessment := range clinvarMatch.Assessments {
				diseases = append(diseases, assessment.ReportedPhenotypeInfo)
				sigs = append(sigs, assessment.ClinicalSignificance)
				genes = append(genes, assessment.SubmittedGeneSymbol)
			}

			rsid := getRsid(line.ID, clinvarMatch.Variant.Info["RS"])
			snpediaLink := ""
			dbSNPLink := ""
			if rsid != "" {
				snpediaLink = fmt.Sprintf(SnpediaLinkPattern, rsid)
				dbSNPLink = fmt.Sprintf(DbSNPLinkPattern, rsid)
			}

			varEnd := line.Pos + len(line.Ref)
			record := []string{
				line.Chrom,
				strconv.Itoa(line.Pos),
				strconv.Itoa(varEnd),
				clinvarMatch.Variant.Info["CLNVC"],
				line.Qual,
				line.Filter,
				line.Ref,
				line.Alt,
				rsid,
				line.GetSampleData("GT"),
				clinvarMatch.Variant.ID,
				strconv.Itoa(clinvarMatch.AssessmentCount),
				clinvarMatch.Pathogenicity.ToString(),
				strconv.Itoa(clinvarMatch.PathogenicityCounts[clinvar.PathogenicityBenign]),
				strconv.Itoa(clinvarMatch.PathogenicityCounts[clinvar.PathogenicityLikelyBenign]),
				strconv.Itoa(clinvarMatch.PathogenicityCounts[clinvar.PathogenicityVUS]),
				strconv.Itoa(clinvarMatch.PathogenicityCounts[clinvar.PathogenicityLikelyPathogenic]),
				strconv.Itoa(clinvarMatch.PathogenicityCounts[clinvar.PathogenicityPathogenic]),
				strconv.Itoa(clinvarMatch.PathogenicityCounts[clinvar.PathogenicityOther]),
				strings.Join(clinvarMatch.Diseases, ","),
				strings.Join(clinvarMatch.Genes, ","),
				fmt.Sprintf(ClinvarLinkPattern, clinvarMatch.Variant.ID),
				dbSNPLink,
				snpediaLink,
				clinvarMatch.Variant.Info["AF_ESP"],
				clinvarMatch.Variant.Info["AF_EXAC"],
				clinvarMatch.Variant.Info["AF_TGP"],
			}
			err := writer.Write(record)
			if err != nil {
				return err
			}
		}
	}
	writer.Flush()
	log.Infof("Wrote %d assessed variants to %s\n", matches, config.OutputFile)
	return nil
}
