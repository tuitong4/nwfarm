package main

import (
	"fmt"
	"os"
	"net"
)


type NetworkSwitch struct {
	Name      string
	SerailNum string
	MgtAddr   string
}

type LineCard struct {
	Device     *NetworkSwitch
	ChassisNum int
	SlotNum    int
	SubslotNum int
}

type AggregatedInterface struct {
	AggrNum     int
	Description string
}

type PhysicalInterface struct {
	Name        string
	ShortName   string
	AggrGroup   *AggrefateInterface
	PortNum     int
	OperStatus  string
	AdminStatus string
	LineCard    *LineCard
}


func main() {
	fmt.Println("")
}
