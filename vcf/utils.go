package vcf

import "strings"

func parseInfo(info string) map[string]string {
	infoMap := make(map[string]string)
	infoParts := strings.Split(info, ";")
	for _, infoPart := range infoParts {
		parts := strings.Split(infoPart, "=")
		if len(parts) == 2 {
			infoMap[parts[0]] = parts[1]
		} else {
			infoMap[parts[0]] = ""
		}
	}
	return infoMap
}
