package main

import (
/*	"errors"
	"strconv"
	"strings"*/
	"fmt"
	"./net"
)



func main() {
	ip, err := net.IP("172.25.68.1") 

	if err != nil{
		fmt.Println(err.Error())
	}   
	fmt.Println(ip.NetworkAddr().IPStr())

	t := net.TST{"hahaha"}
	fmt.Println(t.Name)

}
