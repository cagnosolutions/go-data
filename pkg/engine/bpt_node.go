package engine

type Node = Page

type BTree = Node

func startNewTree(k [8]byte, ptr *Record) *Node {
	return nil
}

func (n *Node) setIsLeaf(isSet bool) {
	if isSet {
		n.setFlags(P_LEAF)
		return
	}
	n.setFlags(P_NODE)
}

func (n *Node) getIsLeaf() bool {
	return n.hasFlag(P_LEAF)
}
