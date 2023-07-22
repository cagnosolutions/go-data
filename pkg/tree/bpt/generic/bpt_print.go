package generic

import (
	"fmt"
	"strings"
)

func newQueue[K Key, V any]() *print[K, V] {
	return new(print[K, V])
}

type print[K Key, V any] struct {
	node *node[K, V]
	next *print[K, V]
}

// String is node's stringer method
func (n *print[K, V]) String() string {
	ss := fmt.Sprintf("\tr%dn%d[", height(n.node), pathToRoot(n.node.parent, n.node))
	for i := 0; i < n.node.numKeys-1; i++ {
		ss += fmt.Sprintf("%v", n.node.keys[i])
		ss += fmt.Sprintf(",")
	}
	ss += fmt.Sprintf("%v]", n.node.keys[n.node.numKeys-1])
	return ss
}

func nodeID[K Key, V any](n *node[K, V]) string {
	ss := fmt.Sprintf("h%.4xk", height(n))
	for i := 0; i < n.numKeys-1; i++ {
		ss += fmt.Sprintf("%v", n.keys[i])
	}
	ss += fmt.Sprintf("%v", n.keys[n.numKeys-1])
	return ss
}

func printNodeMarkdown[K Key, V any](n *node[K, V]) {
	ss := fmt.Sprintf("\t%s[", nodeID(n))
	for i := 0; i < n.numKeys-1; i++ {
		ss += fmt.Sprintf("%v", n.keys[i])
		ss += fmt.Sprintf(",")
	}
	ss += fmt.Sprintf("%v]", n.keys[n.numKeys-1])
	if !n.isLeaf {
		cc := make([]string, n.numKeys)
		for i := 0; i <= n.numKeys; i++ {
			child := (*node[K, V])(n.ptrs[i])
			cc = append(cc, fmt.Sprintf("%s --- %s", ss, nodeID[K, V](child)))
		}
		ss = strings.Join(cc, "\n")
	}
	fmt.Println(ss)
}

func (n *node[K, V]) _String() string {
	ss := fmt.Sprintf("[")
	for i := 0; i < n.numKeys-1; i++ {
		ss += fmt.Sprintf("%v|", n.keys[i])
	}
	ss += fmt.Sprintf("%v]", n.keys[n.numKeys-1])
	return ss
}

func newPrint[K Key, V any](n *node[K, V]) *print[K, V] {
	return &print[K, V]{
		node: n,
		next: nil,
	}
}

func enqueue[K Key, V any](queue *print[K, V], newNode *node[K, V]) *print[K, V] {
	var c *print[K, V]
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
	return queue
}

func dequeue[K Key, V any](queue *print[K, V]) (*print[K, V], *print[K, V]) {
	var n *print[K, V]
	n = queue
	queue = queue.next
	n.next = nil
	return n, queue
}

func printLeaves[K Key, V any](root *node[K, V]) {
	if root == nil {
		fmt.Println("empty tree")
		return
	}
	var c *node[K, V]
	c = root
	for !c.isLeaf {
		c = (*node[K, V])(c.ptrs[0])
	}
	for {
		for i := 0; i < c.numKeys; i++ {
			fmt.Printf("%v ", c.keys[i])
			fmt.Printf("%p ", c.ptrs[i])
		}
		fmt.Printf("%p ", c.ptrs[order-1])
		if c.ptrs[order-1] != nil {
			fmt.Printf(" | ")
			c = (*node[K, V])(c.ptrs[order-1])
		} else {
			break
		}
		fmt.Printf("\n")
	}
}

// height is a utility function to give the height of the tree, which
// length in number of edges of the path from the root to any leaf
func height[K Key, V any](root *node[K, V]) int {
	h := 0
	var c *node[K, V]
	c = root
	for !c.isLeaf {
		c = (*node[K, V])(c.ptrs[0])
		h++
	}
	return h
}

// pathToRoot is a utility function to give the length in edges of
// the path from any node to the root
func pathToRoot[K Key, V any](root *node[K, V], child *node[K, V]) int {
	length := 0
	var c *node[K, V]
	c = child
	for c != root {
		c = c.parent
		length++
	}
	return length
}
func printTree[K Key, V any](root *node[K, V]) {
	var rank, newRank int
	if root == nil {
		fmt.Println("empty tree")
		return
	}
	var queue *print[K, V]
	queue = nil
	queue = enqueue[K, V](queue, root)
	fmt.Println("graph TD")
	fmt.Printf("\ttitle{B+Tree of order %d}\n", M)
	for queue != nil {
		var n *print[K, V]
		n, queue = dequeue[K, V](queue)
		if n.node.parent != nil && n.node == (*node[K, V])(n.node.parent.ptrs[0]) {
			newRank = pathToRoot[K, V](root, n.node)
			if newRank != rank {
				rank = newRank
			}
		}
		printNodeMarkdown(n.node)
		if !n.node.isLeaf {
			for i := 0; i <= n.node.numKeys; i++ {
				queue = enqueue(queue, (*node[K, V])(n.node.ptrs[i]))
			}
		}
	}
}

func _printTree[K Key, V any](root *node[K, V]) {
	var rank, newRank int
	if root == nil {
		fmt.Println("empty tree")
		return
	}
	var queue *print[K, V]
	queue = nil
	// put the root node in the current node
	queue = enqueue[K, V](queue, root)
	var nn int
	for queue != nil {
		// get the current node out of the queue
		var n *print[K, V]
		n, queue = dequeue[K, V](queue)
		if n.node.parent != nil && n.node == (*node[K, V])(n.node.parent.ptrs[0]) {
			newRank = pathToRoot(root, n.node)
			if newRank != rank {
				// fmt.Printf(" (level=%d)", rank)
				rank = newRank
				nn = 0
				fmt.Printf("\n")
			}
		}
		fmt.Printf("\tr%dn%d[", rank, nn)
		for i := 0; i < n.node.numKeys-1; i++ {
			fmt.Printf("%v", n.node.keys[i])
			fmt.Printf(",")
		}
		fmt.Printf("%v]", n.node.keys[n.node.numKeys-1])
		fmt.Printf(" --> ")
		fmt.Printf("\n")
		nn++
		// if not a leaf, queue up the child pointers
		if !n.node.isLeaf {
			for i := 0; i <= n.node.numKeys; i++ {
				child := (*node[K, V])(n.node.ptrs[i])
				queue = enqueue[K, V](queue, child)
			}
		}
		// if it is a leaf, print the values
		// if n.node.isLeaf {
		// 	fmt.Printf("%s", (*node)(n.node.ptrs[order-1]))
		// } else {
		// 	fmt.Printf("%s", (*node)(n.node.ptrs[n.node.numKeys]))
		// }
		// fmt.Printf(" | ")
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

func print_tree[K Key, V any](root *node[K, V]) {
	fmt.Println("Printing Tree...")
	var i, rank, new_rank int
	if root == nil {
		fmt.Printf("Empty tree.\n")
		return
	}
	var queue *print[K, V]
	queue = nil
	queue = enqueue[K, V](queue, root)
	for queue != nil {
		var prt *print[K, V]
		prt, queue = dequeue[K, V](queue)
		if prt.node.parent != nil && prt.node == (*node[K, V])(prt.node.parent.ptrs[0]) {
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
			fmt.Printf("%v|", prt.node.keys[i])
		}
		fmt.Printf("%v]", prt.node.keys[prt.node.numKeys-1])
		if !prt.node.isLeaf {
			for i = 0; i <= prt.node.numKeys; i++ {
				queue = enqueue(queue, (*node[K, V])(prt.node.ptrs[i]))
			}
		}
		fmt.Printf("  ")
	}
	fmt.Printf("\n\n")
}

func print_tree_v2[K Key, V any](root *node[K, V]) {
	fmt.Println("Printing Tree...")
	var i, rank, new_rank int
	if root == nil {
		fmt.Printf("Empty tree.\n")
		return
	}
	var queue *print[K, V]
	queue = nil
	queue = enqueue[K, V](queue, root)
	for queue != nil {
		var prt *print[K, V]
		prt, queue = dequeue[K, V](queue)
		if prt.node.parent != nil && prt.node == (*node[K, V])(prt.node.parent.ptrs[0]) {
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
			fmt.Printf("%v|", prt.node.keys[i])
		}
		fmt.Printf("%v]", prt.node.keys[prt.node.numKeys-1])
		if !prt.node.isLeaf {
			for i = 0; i <= prt.node.numKeys; i++ {
				queue = enqueue(queue, (*node[K, V])(prt.node.ptrs[i]))
			}
		}
		fmt.Printf("  ")
	}
	fmt.Printf("\n\n")
}

func print_markdown_tree[K Key, V any](root *node[K, V]) {
	var sss [][]string
	var i, rank, new_rank int
	if root == nil {
		sss = append(sss, []string{"root[ ]"})
		return
	}
	var queue *print[K, V]
	queue = nil
	queue = enqueue[K, V](queue, root)
	for queue != nil {
		var ss []string
		var prt *print[K, V]
		prt, queue = dequeue[K, V](queue)
		if prt.node.parent != nil && prt.node == (*node[K, V])(prt.node.parent.ptrs[0]) {
			new_rank = pathToRoot(root, prt.node)
			if new_rank != rank {
				rank = new_rank
				// fmt.Printf("\n%s (rank=%d)", ident[rank], rank)
				// ss = append(ss, fmt.Sprintf("r%dn", rank))
			}
		}
		if rank == 0 {
			// fmt.Printf("%s", ident[rank])
			ss = append(ss, fmt.Sprintf("r%dn[%v]", rank, prt.node.keys[i]))
		}
		// fmt.Printf("[")
		for i = 0; i < prt.node.numKeys-1; i++ {
			// fmt.Printf("%d|", prt.node.keys[i].data)
			ss = append(ss, fmt.Sprintf("r%dn[%v]", rank, prt.node.keys[i]))
		}
		// fmt.Printf("%d]", prt.node.keys[prt.node.numKeys-1].data)
		ss = append(ss, fmt.Sprintf("r%dn[%v]**", rank, prt.node.keys[prt.node.numKeys-1]))
		if !prt.node.isLeaf {
			for i = 0; i <= prt.node.numKeys; i++ {
				queue = enqueue(queue, (*node[K, V])(prt.node.ptrs[i]))
			}
		}
		sss = append(sss, ss)
		// fmt.Printf("  ")
	}
	// fmt.Printf("\n\n")

	for i := range sss {
		for j := range sss[i] {
			fmt.Printf("[%s] ", sss[i][j])
		}
		fmt.Printf("\n")
	}
}

func print_leaves[K Key, V any](root *node[K, V]) {
	fmt.Println("Printing Leaves...")
	var i int
	var c = root
	if root == nil {
		fmt.Printf("Empty tree.\n")
		return
	}
	for !c.isLeaf {
		c = (*node[K, V])(c.ptrs[0])
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
			if r := (*record[K, V])(c.ptrs[i]); r == nil {
				fmt.Printf("___, ")
				continue
			} else {
				fmt.Printf("%v ", r.Value)
			}
		}
		if c.ptrs[M-1] != nil {
			fmt.Printf(" || ")
			c = (*node[K, V])(c.ptrs[M-1])
		} else {
			break
		}
	}
	fmt.Printf("\n\n")
}
