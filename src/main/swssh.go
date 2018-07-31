package main

import "strings"

var VENDOR_FLAG = map[string]string{
	"H3C":    "H3C Comware Software",
	"HUAWEI": "Huawei Technologies",
	"NEXUS":  "Cisco Nexus",
	"CISCO":  "Cisco IOS Software",
	"RUIJIE": "Ruijie Networks"}


func detectVendor(s string) string {
	for vendor, keywords := range VENDOR_FLAG{
		if strings.Contains(s, keywords){
			return vendor
		}
	}
	return ""
}


