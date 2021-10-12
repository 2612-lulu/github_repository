package merkletree

import (
	"crypto/sha256"
)

// 默克尔数结构
type MerkleTree struct {
	RootNode *MerkleNode
}

// 默克尔数节点结构
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte // 默克尔树根节点
}

// NewMerkleTree，将节点组建为树
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode
	// 确保必须为2的整数倍节点
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	// 两层循环完成节点树形构造
	for i := 0; i < len(data)/2; i++ {
		var new_level []MerkleNode
		// i=0时，叶节点hash合并
		// i=1时，注意nodes已经不是原来的nodes
		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			new_level = append(new_level, *node)
		}
		// nodes已经升级为此前循环生成的新节点
		nodes = new_level
	}

	mTree := MerkleTree{&nodes[0]} // 构造默克尔树

	return &mTree
}

// NewMerkleNode，创建默克尔树节点，既要支持中间节点，也要支持叶子节点
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	m_node := MerkleNode{}

	// 如果左子节点或右子节点为空，代表其对应的data是最原始数据节点
	if left == nil && right == nil { // 如果是原始数据节点
		hash := sha256.Sum256(data) // 计算原始数据hash
		m_node.Data = hash[:]       // 将[32]byte转化为[]byte
	} else { // 如果不是最原始数据节点
		prev_hashes := append(left.Data, right.Data...) // 将左右子树的数据集合到一起
		hash := sha256.Sum256(prev_hashes)              // 计算左右子树数据hash
		m_node.Data = hash[:]                           // 将[32]byte转化为[]byte
	}
	// 左右子树赋值
	m_node.Left = left
	m_node.Right = right

	return &m_node
}
