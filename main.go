package main

import (
	"github.com/kazmiekr/clinvar-matcher/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
	cmd.Execute()
}
