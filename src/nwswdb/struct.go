package nwswdb

import (
	"nwnet"
)

type NetworkDevice struct {
	Name         string
	MemberDevice []*HardDevice
	MgtAddr      string
	ObbMgtAddr   string
	MemberVlanIf []*VlanInterface
	MemberAggr   []*AggregatedInterface
	Vendor       string
}

type HardDevice struct {
	SerailNum      string
	ChassisNum     int8
	MemberPower    []*PowerModule
	MemberFan      []*FanModule
	MemberLineCard []*LineCard
}

type LineCard struct {
	Name       string
	SerailNum  string
	SlotNum    int
	SubslotNum int
	MemberPort []*PhysicalInterface
}

type PowerModule struct {
	Name       string
	SerailNum  string
	SlotNum    int
	SubslotNum int
}

type FanModule struct {
	Name       string
	SerailNum  string
	SlotNum    int
	SubslotNum int
}

type AggregatedInterface struct {
	AggrNum     int
	Description string
	ShortName   string
	OperStatus  bool
	AdminStatus bool
	MemberPort  []*PhysicalInterface
}

type PhysicalInterface struct {
	Name        string
	Description string
	ShortName   string
	PortNum     int8
	OperStatus  bool
	AdminStatus bool
}

type VlanInterface struct {
	Name        string
	Description string
	ShortName   string
	PortNum     int8
	OperStatus  bool
	AdminStatus bool
}

type BGPSTATUS uint8

const (
	Idle BGPSTATUS = 1 + iota
	Connect
	Active
	OpenSent
	OpenConfirm
	Established
)

type BGP struct {
	ASNum     int
	RouterID  string
	PeerGroup []*BGPPeerGroup
	Peers     []*BGPPeer
}

type BGPPeerGroup struct {
	Name         string
	Type         string
	ConnectIf    string
	Status       BGPSTATUS
	ImportPolicy []string
	ExportPolicy []string
}

type BGPPeer struct {
	Address      nwnet.V4Prefix
	Enabled      bool
	ConnectIf    string
	Status       BGPSTATUS
	ImportPolicy []string
	ExportPolicy []string
	PeerGroup    *BGPPeerGroup
}

type RoutePolicy struct {
	Filter interface{}
	Action interface{}
}

type ACL struct {
	Name      string
	AclNum    int
	Nodes     []*ACLNode
	IndexStep int
}

type ACLNode struct {
	Index      int
	Permitted  bool
	Protocol   string
	SrcAddress nwnet.V4Prefix
	SrcPort    string
	DstAddress nwnet.V4Prefix
	DstPort    string
	VRF        string
}

type IPPrefix struct {
	Name  string
	Nodes []*IPPrefixNode
}

type IPPrefixNode struct {
	Index      int
	Permitted  bool
	Address    nwnet.V4Prefix
	MinMaskLen uint8
	MaxMaskLen uint8
}

type ASPath struct {
	Name  string
	Nodes []*ASPathNode
}

type ASPathNode struct {
	Index     int
	Permitted bool
	ASNExpr   string
}

type Community struct {
	Community string
}

type ROUTETYPE uint8

const (
	EXT_TYPE1 ROUTETYPE = 1 + iota
	EXT_TYPE2
	EXT_TYPE12
	INTERNAL
	ISIS_LEVEL1
	ISIS_LEVEL2
	NNSA_EXT_TYPE1
	NNSA_EXT_TYPE2
	NNSA_EXT_TYPE12
)

type RouteType struct {
	Type ROUTETYPE
}

type OSPF struct {
	RouterID        string
	ProcessID       int8
	AreaID          string
	ActiveNetwork   []*nwnet.V4Prefix
	ActiveInterface []*interface{}
	AggrNetwork     []*AggregatedNetwork
	Nssa            bool
	ImportPolicy    []*RouteRedistribution
}

type RouteRedistribution struct {
	RouteProtocol string
	RouteProcess  uint8
	RouteCost     int
	RouteType     ROUTETYPE
	RoutePolicy   *RoutePolicy
}

type AggregatedNetwork struct {
	Network          *nwnet.V4Prefix
	DetailSuppressed bool
}

type AbrSummaryNetwork struct {
	Network    *nwnet.V4Prefix
	Advertised bool
	Cost       int
}

type AsbrSummaryNetwork struct {
	Network    *nwnet.V4Prefix
	Advertised bool
	Cost       int
}
