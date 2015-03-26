package merkle

import (
	"hash"
	"log"
)

// NewHash provides a hash.Hash to generate a merkle.Tree checksum, given a
// HashMaker for the checksums of the blocks written and the blockSize of each
// block per node in the tree.
func NewHash(hm HashMaker, merkleBlockLength int) hash.Hash {
	mh := new(merkleHash)
	mh.blockSize = merkleBlockLength
	mh.hm = hm
	mh.tree = &Tree{Nodes: []*Node{}, BlockLength: merkleBlockLength}
	mh.lastBlock = make([]byte, merkleBlockLength)
	return mh
}

// TODO make a similar hash.Hash, that accepts an argument of a merkle.Tree,
// that will validate nodes as the new bytes are written. If a new written
// block fails checksum, then return an error on the io.Writer

// TODO satisfy the hash.Hash interface
type merkleHash struct {
	blockSize    int
	tree         *Tree
	hm           HashMaker
	lastBlock    []byte // as needed, for Sum()
	lastBlockLen int
	//partialLastBlock bool
}

// XXX this will be tricky, as the last block can be less than the BlockSize.
// if they get the sum, it will be mh.tree.Root().Checksum() at that point.
//
// But if they continue writing, it would mean a continuation of the bytes in
// the last block. So popping the last node, and having a buffer for the bytes
// in that last partial block.
//
// if that last block was complete, then no worries. start the next node.
func (mh *merkleHash) Sum(b []byte) []byte {
	// TODO check if len(mh.lastBlock) < blockSize
	sum, err := mh.tree.Root().Checksum()
	if err != nil {
		log.Println(err)
	}
	return sum
}

func (mh *merkleHash) Write(b []byte) (int, error) {
	// basically we need to:
	// * include prior partial lastBlock, if any
	// * chunk these writes into blockSize
	// * create Node of the sum
	// * add the Node to the tree
	// * stash remainder in the mh.lastBlock

	var (
		curBlock       = make([]byte, mh.blockSize)
		numBytes   int = 0
		numWritten int
		offset     int = 0
	)
	if mh.lastBlock != nil && mh.lastBlockLen > 0 {
		numBytes = copy(curBlock[:], mh.lastBlock[:])
		// not adding to numWritten, since these blocks were accounted for in a
		// prior Write()

		// then we'll chunk the front of the incoming bytes
		offset = copy(curBlock[numBytes:], b[:(mh.blockSize-numBytes)])
		n, err := NewNodeHashBlock(mh.hm, curBlock)
		if err != nil {
			// XXX might need to stash again the prior lastBlock and first little
			// chunk
			return numWritten, err
		}
		mh.tree.Nodes = append(mh.tree.Nodes, n)
		numWritten += offset
	}

	numBytes = (len(b) - offset)
	for i := 0; i < numBytes/mh.blockSize; i++ {
		// TODO Node for curBlock
	}

	// TODO stash (numBytes % mh.blockSize) in mh.lastBlock

	// TODO if len(mh.lastBlock) < blockSize, then set that before returning
	return numWritten, nil
}

func (mh *merkleHash) Reset() {
	mh.tree = &Tree{}
	mh.lastBlock = nil
}

// likely not the best to pass this through and not use our own node block
// size, but let's revisit this.
func (mh *merkleHash) BlockSize() int { return mh.hm().BlockSize() }
func (mh *merkleHash) Size() int      { return mh.hm().Size() }
