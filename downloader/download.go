package downloader

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/schollz/progressbar/v3"
)

func DownloadFile(filepath string, url string, description string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		description,
	)
	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	return err
}

func HashFileMd5(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil
}

func ReadFileMd5(md5Path string) (string, error) {
	dat, err := ioutil.ReadFile(md5Path)
	if err != nil {
		return "", err
	}
	md5Contents := string(dat)
	md5Parts := strings.Split(md5Contents, " ")
	md5 := md5Parts[0]
	return md5, nil
}

func VerifyFile(file string, md5File string) error {
	truthMd5, err := ReadFileMd5(md5File)
	if err != nil {
		return err
	}
	localMd5, err := HashFileMd5(file)
	if err != nil {
		return err
	}
	if localMd5 != truthMd5 {
		return fmt.Errorf("md5 mismatch on %s: %s(local) != %s(remote)", file, localMd5, truthMd5)
	}
	return nil
}
