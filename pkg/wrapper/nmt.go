package wrapper

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"sort"

	"github.com/celestiaorg/nmt"
)

// ModuleLeaf represents a single leaf in the Module Merkle Tree.
type ModuleLeaf struct {
	Namespace []byte
	Data      []byte
}

// AppStateNMTWrapper holds the data required to build a Namespace Merkle Tree
// over module commit hashes.
type AppStateNMTWrapper struct {
	BaseNamespace []byte
	Leaves        []ModuleLeaf
	Tree          *nmt.NamespacedMerkleTree
}

// ComputeNamespace calculates an 8-byte namespace identifier for a module using
// sha256(base || moduleName) and truncating the result to 8 bytes.
func ComputeNamespace(base []byte, moduleName string) []byte {
	h := sha256.Sum256(append(base, []byte(moduleName)...))
	return h[:8]
}

// AddModule adds a module and its commit hash as a leaf in the tree.
func (w *AppStateNMTWrapper) AddModule(moduleName string, commitHash []byte) {
	ns := ComputeNamespace(w.BaseNamespace, moduleName)
	w.Leaves = append(w.Leaves, ModuleLeaf{Namespace: ns, Data: commitHash})
}

// Build constructs the NMT from the collected module leaves and returns the
// resulting root hash.
func (w *AppStateNMTWrapper) Build() ([]byte, error) {
	if len(w.Leaves) == 0 {
		return nil, errors.New("no module leaves to build tree")
	}

	sort.Slice(w.Leaves, func(i, j int) bool {
		return bytes.Compare(w.Leaves[i].Namespace, w.Leaves[j].Namespace) < 0
	})

	tree := nmt.New(sha256.New, nmt.NamespaceIDSize(8))
	for _, leaf := range w.Leaves {
		data := append(append([]byte{}, leaf.Namespace...), leaf.Data...)
		if err := tree.Push(data); err != nil {
			return nil, err
		}
	}

	w.Tree = tree
	return tree.Root(), nil
}

// GetInclusionProof returns the inclusion proof for the specified namespace.
func (w *AppStateNMTWrapper) GetInclusionProof(ns []byte) (*nmt.Proof, error) {
	if w.Tree == nil {
		return nil, errors.New("tree not built")
	}
	return w.Tree.ProveNamespace(ns)
}
