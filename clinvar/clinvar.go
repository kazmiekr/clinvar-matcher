package clinvar

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/kazmiekr/clinvar-matcher/vcf"
	log "github.com/sirupsen/logrus"
)

const (
	SignficanceKey           = "CLNSIG"
	RSIDKey                  = "RS"
	VariantClassificationKey = "CLNVC"
	AFEspKey                 = "AF_ESP"
	AFExacKey                = "AF_EXAC"
	AFTgpKey                 = "AF_TGP"
)

type Pathogenicity int

const (
	PathogenicityBenign Pathogenicity = iota + 1
	PathogenicityLikelyBenign
	PathogenicityVUS
	PathogenicityLikelyPathogenic
	PathogenicityPathogenic
	PathogenicityOther
)

var PathogenicitytoString = map[Pathogenicity]string{
	PathogenicityBenign:           "Benign",
	PathogenicityLikelyBenign:     "Likely Benign",
	PathogenicityVUS:              "VUS",
	PathogenicityLikelyPathogenic: "Likely Pathogenic",
	PathogenicityPathogenic:       "Pathogenic",
	PathogenicityOther:            "Other",
}

func (p Pathogenicity) ToString() string {
	return PathogenicitytoString[p]
}

type ClinvarClient struct {
	Variants        []*vcf.VcfLine
	VariantsByKey   map[string]*vcf.VcfLine
	Assessments     []*ClinvarSubmission
	AssessmentsByID map[string][]*ClinvarSubmission
}

type ClinvarRecord struct {
	Variant             *vcf.VcfLine
	Assessments         []*ClinvarSubmission
	Pathogenicity       Pathogenicity
	PathogenicityCounts map[Pathogenicity]int
	AssessmentCount     int
	Diseases            []string
	Genes               []string
}

type ClinvarSubmission struct {
	VariationID                 string
	ClinicalSignificance        string
	DateLastEvaluated           string
	Description                 string
	SubmittedPhenotypeInfo      string
	ReportedPhenotypeInfo       string
	ReviewStatus                string
	CollectionMethod            string
	OriginCounts                string
	Submitter                   string
	SCV                         string
	SubmittedGeneSymbol         string
	ExplanationOfInterpretation string
	Pathogenicity               Pathogenicity
	Disease                     ClinvarDisease
}

type ClinvarDisease struct {
	MedGenID    string
	DiseaseName string
}

func (clinvar *ClinvarClient) Lookup(variantClinvarKey string) (*ClinvarRecord, bool) {
	variant, ok := clinvar.VariantsByKey[variantClinvarKey]
	if !ok {
		return nil, ok
	}
	assessments, ok := clinvar.AssessmentsByID[variant.ID]
	if !ok {
		return nil, ok
	}

	pathogenicityCounts := make(map[Pathogenicity]int)
	diseases := make(map[string]struct{})
	genes := make(map[string]struct{})
	maxPathogenicity := PathogenicityBenign

	allPaths := []Pathogenicity{PathogenicityBenign, PathogenicityLikelyBenign, PathogenicityVUS, PathogenicityLikelyPathogenic, PathogenicityPathogenic, PathogenicityOther}
	for _, p := range allPaths {
		pathogenicityCounts[p] = 0
	}

	for _, assessment := range assessments {
		pathogenicityCounts[assessment.Pathogenicity]++
		if assessment.Pathogenicity > maxPathogenicity {
			maxPathogenicity = assessment.Pathogenicity
		}
		if _, ok := diseases[assessment.Disease.DiseaseName]; !ok {
			diseases[assessment.Disease.DiseaseName] = struct{}{}
		}
		if assessment.SubmittedGeneSymbol != "-" {
			if _, ok := genes[assessment.SubmittedGeneSymbol]; !ok {
				genes[assessment.SubmittedGeneSymbol] = struct{}{}
			}
		}
	}

	uniqueDiseases := make([]string, 0)
	for disease := range diseases {
		uniqueDiseases = append(uniqueDiseases, disease)
	}
	sort.Strings(uniqueDiseases)

	uniqueGenes := make([]string, 0)
	for gene := range genes {
		uniqueGenes = append(uniqueGenes, gene)
	}
	sort.Strings(uniqueGenes)

	return &ClinvarRecord{
		Variant:             variant,
		Assessments:         assessments,
		AssessmentCount:     len(assessments),
		PathogenicityCounts: pathogenicityCounts,
		Pathogenicity:       maxPathogenicity,
		Diseases:            uniqueDiseases,
		Genes:               uniqueGenes,
	}, ok
}

func (clinvar *ClinvarClient) PrintPathogenicityStats() {
	submissionMap := make(map[string]int)
	for _, assessment := range clinvar.Assessments {
		p := assessment.ClinicalSignificance
		if _, ok := submissionMap[p]; !ok {
			submissionMap[p] = 0
		}
		submissionMap[p]++
	}
	fmt.Println("-----")
	for key, value := range submissionMap {
		fmt.Printf("%v,%d\n", key, value)
	}
}

func ToClinvarKey(vcfLine *vcf.VcfLine) string {
	return fmt.Sprintf("%s:%d %s:%s", vcfLine.Chrom, vcfLine.Pos, vcfLine.Ref, vcfLine.Alt)
}

type assessmentLoad struct {
	assessments []*vcf.VcfLine
	err         error
}

func loadAssessments(assessmentsFile string, loadAssessmentsChan chan assessmentLoad, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Infof("Loading clinvar assessments from %s", assessmentsFile)
	assessments, err := vcf.ReadVcf(assessmentsFile)
	loadAssessmentsChan <- assessmentLoad{
		assessments: assessments,
		err:         err,
	}
}

type submissionLoad struct {
	submissions []*ClinvarSubmission
	err         error
}

func loadSubmissions(submissionFile string, loadSubmissionsChan chan submissionLoad, wg *sync.WaitGroup) {
	defer wg.Done()
	submissions, err := ParseSubmissionSummary(submissionFile)
	loadSubmissionsChan <- submissionLoad{
		submissions: submissions,
		err:         err,
	}
}

func getDiseaseFromPhenotype(phenotypeInfo string) ClinvarDisease {
	parts := strings.Split(phenotypeInfo, ":")
	if len(parts) != 2 {
		return ClinvarDisease{}
	}
	return ClinvarDisease{
		MedGenID:    parts[0],
		DiseaseName: parts[1],
	}
}

func getPathogencityFromClinsig(clinsig string) Pathogenicity {
	pathogenicities := strings.Split(clinsig, ",")
	p := strings.ToLower(pathogenicities[0])
	p = strings.Replace(p, "_", " ", -1)

	switch p {
	case "benign":
		return PathogenicityBenign
	case "benign/likely benign":
		return PathogenicityLikelyBenign
	case "likely benign":
		return PathogenicityLikelyBenign
	case "uncertain significance":
		return PathogenicityVUS
	case "likely pathogenic":
		return PathogenicityLikelyPathogenic
	case "pathogenic/likely pathogenic":
		return PathogenicityLikelyPathogenic
	case "pathogenic":
		return PathogenicityPathogenic
	default:
		return PathogenicityOther
	}
}

func NewClinvar(assessmentsFile string, submissionFile string) (*ClinvarClient, error) {
	clinvarClient := &ClinvarClient{}

	loadAssessmentsChan := make(chan assessmentLoad, 1)
	loadSubmissionsChan := make(chan submissionLoad, 1)

	var wg sync.WaitGroup
	wg.Add(2)
	go loadAssessments(assessmentsFile, loadAssessmentsChan, &wg)
	go loadSubmissions(submissionFile, loadSubmissionsChan, &wg)
	wg.Wait()

	loadAssessmentsResult := <-loadAssessmentsChan
	if loadAssessmentsResult.err != nil {
		return clinvarClient, loadAssessmentsResult.err
	}

	log.Infof("Clinvar Assessment Count: %d\n", len(loadAssessmentsResult.assessments))
	clinvarClient.VariantsByKey = buildClinvarMap(loadAssessmentsResult.assessments)

	loadSubmissionsResult := <-loadSubmissionsChan
	if loadSubmissionsResult.err != nil {
		return clinvarClient, loadSubmissionsResult.err
	}
	log.Infof("Clinvar Submission Count: %d\n", len(loadSubmissionsResult.submissions))

	submissionsMap := make(map[string][]*ClinvarSubmission)
	for _, submission := range loadSubmissionsResult.submissions {
		if _, ok := submissionsMap[submission.VariationID]; !ok {
			submissionsMap[submission.VariationID] = make([]*ClinvarSubmission, 0)
		}
		submissionsMap[submission.VariationID] = append(submissionsMap[submission.VariationID], submission)
	}
	clinvarClient.AssessmentsByID = submissionsMap

	return clinvarClient, nil
}

func buildClinvarMap(lines []*vcf.VcfLine) map[string]*vcf.VcfLine {
	clinvarMap := make(map[string]*vcf.VcfLine)
	for _, line := range lines {
		clinvarMap[ToClinvarKey(line)] = line
	}
	return clinvarMap
}

func ParseSubmissionSummary(filePath string) ([]*ClinvarSubmission, error) {
	log.Infof("Loading clinvar submission summary from %s", filePath)
	file, err := os.Open(filePath)
	submissions := make([]*ClinvarSubmission, 0)
	if err != nil {
		return submissions, err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return submissions, err
	}
	defer gz.Close()
	scanner := bufio.NewScanner(gz)

	for scanner.Scan() {
		line := scanner.Text()
		if line[0] == '#' {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 13 {
			continue
		}
		reportedPhenotype := parts[5]
		submission := &ClinvarSubmission{
			VariationID:                 parts[0],
			ClinicalSignificance:        parts[1],
			Pathogenicity:               getPathogencityFromClinsig(parts[1]),
			Disease:                     getDiseaseFromPhenotype(reportedPhenotype),
			DateLastEvaluated:           parts[2],
			Description:                 parts[3],
			SubmittedPhenotypeInfo:      parts[4],
			ReportedPhenotypeInfo:       reportedPhenotype,
			ReviewStatus:                parts[6],
			CollectionMethod:            parts[7],
			OriginCounts:                parts[8],
			Submitter:                   parts[9],
			SCV:                         parts[10],
			SubmittedGeneSymbol:         parts[11],
			ExplanationOfInterpretation: parts[12],
		}
		submissions = append(submissions, submission)
	}

	if err := scanner.Err(); err != nil {
		return submissions, err
	}
	return submissions, nil
}
