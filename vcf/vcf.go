package vcf

import (
	"archive/zip"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type VcfLine struct {
	Chrom      string
	Pos        int
	ID         string
	Ref        string
	Alt        string
	Qual       string
	Filter     string
	Info       map[string]string
	Format     string
	Sample     string
	SampleData map[string]string
}

func (vcfLine VcfLine) GetSampleData(key string) string {
	return vcfLine.SampleData[key]
}

func ReadVcf(vcfPath string) ([]*VcfLine, error) {
	lines := make([]*VcfLine, 0)
	var reader io.Reader
	if strings.HasSuffix(vcfPath, "vcf.gz") {
		file, err := os.Open(vcfPath)
		if err != nil {
			return lines, err
		}
		defer file.Close()
		gz, err := gzip.NewReader(file)
		if err != nil {
			return lines, err
		}
		defer gz.Close()
		reader = gz
	} else if strings.HasSuffix(vcfPath, "vcf.zip") {
		zipReader, err := zip.OpenReader(vcfPath)
		if err != nil {
			return lines, err
		}
		defer zipReader.Close()
		// Ensure the zip only has one file and it's a VCF file
		if len(zipReader.File) == 1 && strings.HasSuffix(zipReader.File[0].Name, "vcf") {
			file, err := zipReader.File[0].Open()
			if err != nil {
				return lines, err
			}
			defer file.Close()
			reader = file
		} else {
			return lines, fmt.Errorf("unable to locate vcf in zip")
		}
	} else if strings.HasSuffix(vcfPath, "vcf") {
		file, err := os.Open(vcfPath)
		if err != nil {
			return lines, err
		}
		defer file.Close()
		reader = file
	} else {
		return lines, fmt.Errorf("please supply a .vcf or .vcf.gz to read")
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		vcfLine, err := parseVcfLine(line)
		if err != nil {
			return lines, err
		}
		if vcfLine != nil {
			lines = append(lines, vcfLine)
		}
	}

	if err := scanner.Err(); err != nil {
		return lines, err
	}
	return lines, nil
}

func parseVcfLine(line string) (*VcfLine, error) {
	if line[0] == '#' {
		return nil, nil
	}
	parts := strings.Split(line, "\t")
	if len(parts) < 7 {
		return nil, nil
	}
	pos, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}
	sample := ""
	format := ""
	sampleData := make(map[string]string)
	if len(parts) == 10 {
		format = parts[8]
		sample = parts[9]
		formats := strings.Split(format, ":")
		samples := strings.Split(sample, ":")
		if len(formats) == len(samples) {
			for x := 0; x < len(formats); x++ {
				sampleData[formats[x]] = samples[x]
			}
		} else {
			variantKey := fmt.Sprintf("%v:%d", parts[0], pos)
			log.Warnf("Format/Sample length mismatch for variant: %v", variantKey)
		}
	}

	return &VcfLine{
		Chrom:      parts[0],
		Pos:        pos,
		ID:         parts[2],
		Ref:        parts[3],
		Alt:        parts[4],
		Qual:       parts[5],
		Filter:     parts[6],
		Info:       parseInfo(parts[7]),
		Format:     format,
		Sample:     sample,
		SampleData: sampleData,
	}, nil
}
