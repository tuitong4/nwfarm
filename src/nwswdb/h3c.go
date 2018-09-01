package nwswdb

import "strings"

func GetInterfaceSection(config []string) ([]string, error) {
	var start, end int = 0, 0
	matched := false
	configlen := len(config)
	for i, line := range config {
		if strings.HasPrefix(line, "interface") {
			start = i
			matched = true
			continue
		}

		if strings.HasPrefix(line, "#") && matched {
			if i >= configlen {
				return config[start:end]
			}
		}
	}
}
