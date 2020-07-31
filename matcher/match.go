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

// Will return the clinvar rsid first, otherwise it'll try to use the vcf id
func getRsid(vcfFileId, clinvarRsid string) string {
	if clinvarRsid != "" {
		return fmt.Sprintf("rs%s", clinvarRsid)
	}
	return vcfFileId
}

func GenerateAssessmentReport(vcfFile string, clinvarFileParam string, clinvarSubmissionsFileParam string, outputFileParam string) error {

	// If user did not specify clinvar VCF file, download latest
	clinvarFile := clinvarFileParam
	if strings.Index(clinvarFileParam, "https://") != -1 {
		description := "Downloading Clinvar VCF"
		localFile := path.Base(clinvarFileParam)
		err := downloader.DownloadFile(localFile, clinvarFileParam, description)
		if err != nil {
			return err
		}
		clinvarFile = localFile
	}

	// If user did not specify clinvar submissions file, download latest
	clinvarSubmissionFile := clinvarSubmissionsFileParam
	if strings.Index(clinvarSubmissionsFileParam, "https://") != -1 {
		description := "Downloading Clinvar Submissions"
		localFile := path.Base(clinvarSubmissionsFileParam)
		err := downloader.DownloadFile(localFile, clinvarSubmissionsFileParam, description)
		if err != nil {
			return err
		}
		clinvarSubmissionFile = localFile
	}

	return WriteAssessedVariants(vcfFile, clinvarFile, clinvarSubmissionFile, outputFileParam)
}

func WriteAssessedVariants(vcfFile string, clinvarFile string, clinvarSubmissionsFile string, outputFile string) error {
	log.Infof("Loading vcf from %s", vcfFile)
	variants, err := vcf.ReadVcf(vcfFile)
	if err != nil {
		return err
	}
	log.Infof("Variant Count: %d\n", len(variants))

	clinvarClient, err := clinvar.NewClinvar(clinvarFile, clinvarSubmissionsFile)
	if err != nil {
		return err
	}

	matches := 0
	counts := make(map[string]int)

	resultFile, err := os.Create(outputFile)
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

	for _, line := range variants {
		// Quality filter
		if strings.ToLower(line.Filter) != "pass" {
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
	log.Infof("Wrote %d assessed variants to %s\n", matches, outputFile)
	return nil
}
