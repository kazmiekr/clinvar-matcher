# Clinvar-Matcher

This is tool that will take in a VCF file and find all the variants that have assessment information from ClinVar. It'll download the latest files when you run it, and will output a CSV file of all your variants with assessments.

Run and download latest

```
./clinvar-matcher my_vcf.vcf
```

If you already have the ClinVar files, you can supply those to avoid extra downloads

```
./clinvar-matcher my_vcf.vcf -s submission_summary.txt.gz -c clinvar.vcf.gz
```

Full details

```
Clinvar Matcher is a tool to match your vcf with the latest Clinvar

Usage:
  clinvar-matcher [vcfFile] [flags]

Flags:
  -s, --clinvar-submissions string   Clinvar submission summary file, leave blank to download latest (default "https://ftp.ncbi.nlm.nih.gov/pub/clinvar/tab_delimited/submission_summary.txt.gz")
  -c, --clinvar-vcf string           Clinvar vcf file, leave blank to download latest (default "https://ftp.ncbi.nlm.nih.gov/pub/clinvar/vcf_GRCh37/clinvar.vcf.gz")
  -h, --help                         help for clinvar-matcher
  -o, --output-file string           Output file to write (default "clinvar_assessments.csv")
```