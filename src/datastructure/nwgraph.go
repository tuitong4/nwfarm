package datastructure


type Link struct {
	remoteID string; //Used as switch's ip address or system mac. If id is confliced, use with a namespace.
	localPort string;
	remotePort string;
}


type Peer struct {
	ID string;  //Used as switch's ip address or system mac. If id is confliced, use with a namespace.
}

type Node struct {
	ID string; //Used as switch's ip address or system mac.
	links []Link;
}


type Set struct {
	ID string;
	nodes []Node;
}

