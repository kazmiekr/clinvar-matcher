source = ["./dist/clinvar-matcher-macos_darwin_amd64/clinvar-matcher"]
bundle_id = "com.kevinkaz.clinvarmatcher"

sign {
  application_identity = "Developer ID Application: Kevin Kazmierczak"
}

zip {
  output_path = "dist/clinvar-matcher-macos_darwin_amd64_signed.zip"
}