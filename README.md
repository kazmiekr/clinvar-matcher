# ClinVar-Matcher

This is tool that will take in a VCF file and find all the variants that have assessment information from ClinVar. It'll download the latest files when you run it, and will output a CSV file of all your variants with assessments.

## How does this work?

This tool will download the latest ClinVar vcf file (default GRCh37) and also the ClinVar Submission summary file and load up that information into memory. Then it'll open your VCF file (you can leave it compressed as gz or zip), iterate over the variants in that file and look up in ClinVar to see if there is a match based on the chromosome, position, reference sequence, and variant sequence. For every match, it'll aggregate the submission information and write a record to the csv file.

By default when it finishes, it'll delete the ClinVar source files, which are about 100mb total.

## Running

Run the latest release for your operating system and extract the zip. Then you can run the command in a terminal like the examples below. You can pass it your VCF directly, or it also supports reading a gzipped VCF or a plain zip that contains only your VCF, without having to extract it.

If you don't want to use the latest files, you can always specify a different version with the `--clinvar-vcf` flag, for example you wanted to use a file from last year.

```
./clinvar-matcher my_vcf.vcf
```

Full details with optional flags:

```
Clinvar Matcher is a tool to match your vcf with the latest ClinVar

Usage:
  clinvar-matcher [vcfFile] [flags]

Flags:
  -s, --clinvar-submissions string   ClinVar submission summary file, leave blank to download latest (default "https://ftp.ncbi.nlm.nih.gov/pub/clinvar/tab_delimited/submission_summary.txt.gz")
  -c, --clinvar-vcf string           ClinVar vcf file, leave blank to download latest (default "https://ftp.ncbi.nlm.nih.gov/pub/clinvar/vcf_GRCh37/clinvar.vcf.gz")
  -h, --help                         help for clinvar-matcher
  -a, --include-all                  Include low quality, non passing variants. Will use PASSing variants by default
  -k, --keep-downloads               Keep the ClinVar downloaded files when complete, will be deleted by default
  -o, --output-file string           Output file to write (default "clinvar_assessments.csv")
```

## Columns in Report CSV

* Chromosome - Chromosome without a 'chr' prefix
* Begin - Begin position reported from VCF
* End - Begin position plus length of `Ref`
* Var Type - Variant type reported from ClinVar
* Quality - Variant `QUAL` reported from VCF
* Filter - Variant `FILTER` reported from VCF
* Ref - Reference sequence from VCF
* Alt - Variant sequence from VCF
* Rsid - Reference SNP ID assigned by dbSNP, will use what's in ClinVar first, then check source VCF for the `ID` field
* Zygosity - Genotype reported from VCF in `GT` format field
* Clinvar ID - ClinVar Variation ID
* Assessment Count - Total number of ClinVar assessments
* Max Pathogenicity - The highest pathogenicity of the assessments, for example, 1 pathogenic, and 2 VUS, would report pathogenic for this variant.
* \# Benign - Total benign assessments
* \# Likely Benign - Total likely benign assessments
* \# VUS - Total VUS assessments
* \# Likely Path - Total likely pathogenic assessments
* \# Pathogenic - Total pathogenic assessments
* \# Other - Total of assessments that don't fit the above classifications
* Diseases - List of unique disease names from `Disease` fields from assessment submissions
* Genes - List of unique `SubmittedGeneSymbol` fields from assessment submissions
* Clinvar Link - Link to ClinVar variant
* dbSNP Link - Link to to variant at dbSNP
* Snpedia Link - Link to variant at Snpedia
* AF_ESP - Allele frequencies from GO-ESP, provided from ClinVar
* AF_EXAC - Allele frequencies from ExAC, provided from ClinVar
* AF_TGP - Allele frequencies from TGP, provided from ClinVar