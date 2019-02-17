package graph

import (
	"nwnet"
	"strconv"
)

type Link struct {
	remoteID   string //Used as switch's ip address or system mac. If id is confliced, use with a namespace.
	localPort  string
	remotePort string
}

type Node struct {
	ID    string //Used as switch's ip address or system mac.
	links []Link
}

type Cluster struct {
	ID      string
	nodes   []Node
	gnodes  []GNode
	gedges  []GEdge
	nodeids map[string]uint32
}

/*Graph Node, convert to visjs node*/
type GNode struct {
	id    uint32
	level int8
	group int8
	lable string
}

/*Graph Edge, convert to visjs edge*/
type GEdge struct {
	id    string //如果nodeA到NodeB的edge，则edge的id由nodeA和nodeB的id相加得到
	from  uint32 //Gnode的ID
	to    uint32 //Gnode的ID
	lable string
	count int8
}

//对于没法将nodeid转换为uint32的，在GLOBAL_CONUTER中选择一个数
var GLOBAL_CONUTER = uint32(1000)

//记录从GLOBAL_CONUTER选择的节点信息
var GLOBAL_NODEIDS = map[string]uint32{}

func genGNodeID(id string) uint32 {
	ipv4, err := iptool.IPv4(id)
	if err != nil {
		GLOBAL_CONUTER += 1
		if _, ok := GLOBAL_NODEIDS[id]; !ok {
			GLOBAL_NODEIDS[id] = GLOBAL_CONUTER
			return GLOBAL_CONUTER
		}
		return GLOBAL_NODEIDS[id]
	}
	return ipv4.Numeric()
}

func genGEdgeID(id1, id2 string) string {
	return strconv.FormatUint(uint64(genGNodeID(id1))+uint64(genGNodeID(id2)), 10)
}

func genLevel(ipaddr string) int8 {
	return 1
}

func genGroup(ipaddr string) int8 {
	return 1
}
func (c *Cluster) RenderGNodes() {
	for _, node := range c.nodes {
		gn_id := genGNodeID(node.ID)
		gn_level := genLevel(node.ID)
		gn_group := genGroup(node.ID)

		gnode := &GNode{
			id:    gn_id,
			level: gn_level,
			group: gn_group,
			lable: node.ID,
		}

		c.nodeids[node.ID] = gn_id

		c.gnodes = append(c.gnodes, *gnode)
	}

}

func (c *Cluster) RenderEdges() []GEdge {
	edges := map[string]*GEdge{}
	for _, node := range c.nodes {
		for _, link := range node.links {
			ge_id := genGEdgeID(node.ID, link.remoteID)
			if _, ok := edges[ge_id]; ok {
				edges[ge_id].count += 1
			} else {
				edges[ge_id] = &GEdge{
					id:    ge_id,
					from:  c.nodeids[node.ID],
					to:    c.nodeids[link.remoteID],
					lable: "",
					count: 1}
			}
		}
	}
	for _, edge := range edges {
		c.gedges = append(c.gedges, *edge)
	}

}
