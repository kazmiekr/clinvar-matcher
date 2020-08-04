# Dev Notes

Collection of notes for me to remember how figured out some things I'd probably forget later

### How to get your provider id for code signing

```
xcrun altool --list-providers -u "me@myself.com" -p "@env:AC_PASSWORD"
```

### Where are the Apple signing credentials

Set as env vars, `AC_USERNAME`, `AC_PASSWORD`, `AC_PROVIDER`


### How to manually run the app signer

```
gon -log-level=info -log-json ./gon.hcl
```