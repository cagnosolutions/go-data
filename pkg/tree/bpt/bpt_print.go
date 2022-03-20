package bpt

import (
	"fmt"
)

var queue *print = nil

type print struct {
	node *node
	next *print
}

func (n *node) String() string {
	ss := fmt.Sprintf("[")
	for i := 0; i < n.numKeys-1; i++ {
		ss += fmt.Sprintf("%d|", n.keys[i].data)
	}
	ss += fmt.Sprintf("%d]", n.keys[n.numKeys-1].data)
	return ss
}

func newPrint(n *node) *print {
	return &print{
		node: n,
		next: nil,
	}
}

func enqueue(newNode *node) {
	var c *print
	if queue == nil {
		queue = newPrint(newNode)
		queue.next = nil
	} else {
		c = queue
		for c.next != nil {
			c = c.next
		}
		c.next = newPrint(newNode)
		// newNode.next = nil
	}
}

func dequeue() *print {
	var n *print
	n = queue
	queue = queue.next
	n.next = nil
	return n
}

func printLeaves(root *node) {
	if root == nil {
		fmt.Println("empty tree")
		return
	}
	var c *node
	c = root
	for !c.isLeaf {
		c = (*node)(c.ptrs[0])
	}
	for {
		for i := 0; i < c.numKeys; i++ {
			fmt.Printf("%d ", c.keys[i])
			fmt.Printf("%p ", c.ptrs[i])
		}
		fmt.Printf("%p ", c.ptrs[order-1])
		if c.ptrs[order-1] != nil {
			fmt.Printf(" | ")
			c = (*node)(c.ptrs[order-1])
		} else {
			break
		}
		fmt.Printf("\n")
	}
}

// height is a utility function to give the height of the tree, which
// length in number of edges of the path from the root to any leaf
func height(root *node) int {
	h := 0
	var c *node
	c = root
	for !c.isLeaf {
		c = (*node)(c.ptrs[0])
		h++
	}
	return h
}

// pathToRoot is a utility function to give the length in edges of
// the path from any node to the root
func pathToRoot(root *node, child *node) int {
	length := 0
	var c *node
	c = child
	for c != root {
		c = c.parent
		length++
	}
	return length
}

func printTree(root *node) {

	var rank, newRank int

	if root == nil {
		fmt.Println("empty tree")
		return
	}
	queue = nil
	enqueue(root)
	for queue != nil {
		n := dequeue()
		if n.node.parent != nil && n.node == (*node)(n.node.parent.ptrs[0]) {
			newRank = pathToRoot(root, n.node)
			if newRank != rank {
				rank = newRank
				fmt.Printf("\n")
			}
		}
		// fmt.Printf("(%s)", n)
		for i := 0; i < n.node.numKeys; i++ {
			fmt.Printf("%s", n.node)
			// fmt.Printf("%p ", n.node.ptrs[i])
		}
		if !n.node.isLeaf {
			for i := 0; i <= n.node.numKeys; i++ {
				enqueue((*node)(n.node.ptrs[i]))
			}
		}
		if n.node.isLeaf {
			fmt.Printf("%s", (*node)(n.node.ptrs[order-1]))
		} else {
			fmt.Printf("%s", (*node)(n.node.ptrs[n.node.numKeys]))
		}
		fmt.Printf(" | ")
	}
	fmt.Printf("\n")
}

var ident = map[int]string{
	0: "\r\t\t\t\t\t\t\t\t\t\t\t\t",
	1: "\r\t\t\t\t\t\t\t\t\t\t\t",
	2: "\r\t\t\t\t\t\t\t\t",
	3: "\r",
	4: "\r",
	5: "\r",
}

func print_tree(root *node) {
	fmt.Println("Printing Tree...")
	var i, rank, new_rank int
	if root == nil {
		fmt.Printf("Empty tree.\n")
		return
	}
	queue = nil
	enqueue(root)
	for queue != nil {
		prt := dequeue()
		if prt.node.parent != nil && prt.node == (*node)(prt.node.parent.ptrs[0]) {
			new_rank = pathToRoot(root, prt.node)
			if new_rank != rank {
				rank = new_rank
				fmt.Printf("\n%s", ident[rank])
			}
		}
		if rank == 0 {
			fmt.Printf("%s", ident[rank])
		}
		fmt.Printf("[")
		for i = 0; i < prt.node.numKeys-1; i++ {
			fmt.Printf("%d|", prt.node.keys[i].data)
		}
		fmt.Printf("%d]", prt.node.keys[prt.node.numKeys-1].data)
		if !prt.node.isLeaf {
			for i = 0; i <= prt.node.numKeys; i++ {
				enqueue((*node)(prt.node.ptrs[i]))
			}
		}
		fmt.Printf("  ")
	}
	fmt.Printf("\n\n")
}

func print_leaves(root *node) {
	fmt.Println("Printing Leaves...")
	var i int
	var c *node = root
	if root == nil {
		fmt.Printf("Empty tree.\n")
		return
	}
	for !c.isLeaf {
		c = (*node)(c.ptrs[0])
	}
	for {
		/*
			for i = 0; i < M-1; i++ {
				if c.keys[i] == nil {
					fmt.Printf("___, ")
					continue
				}
				//fmt.Printf("%s, ", c.keys[i])
				// extract record / value instead
				rec := (*record)(unsafe.Pointer(c.ptrs[i]))
				fmt.Printf("%s, ", rec.val)
			}
			if c.ptrs[M-1] != nil {
				fmt.Printf(" | ")
				c = (*node)(unsafe.Pointer(c.ptrs[M-1]))
			} else {
				break
			}
		*/
		for i = 0; i < M-1; i++ {
			if r := (*record)(c.ptrs[i]); r == nil {
				fmt.Printf("___, ")
				continue
			} else {
				fmt.Printf("%s ", r.Value)
			}
		}
		if c.ptrs[M-1] != nil {
			fmt.Printf(" || ")
			c = (*node)(c.ptrs[M-1])
		} else {
			break
		}
	}
	fmt.Printf("\n\n")
}
