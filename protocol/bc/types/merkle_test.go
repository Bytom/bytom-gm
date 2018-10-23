package types

import (
	"encoding/hex"

	"math/rand"
	"testing"
	"time"

	"github.com/bytom-gm/protocol/bc"
	"github.com/bytom-gm/protocol/vm"
	"github.com/bytom-gm/testutil"
)

func TestMerkleRoot(t *testing.T) {
	cases := []struct {
		witnesses [][][]byte
		want      bc.Hash
	}{{
		witnesses: [][][]byte{
			{
				{1},
				[]byte("00000"),
			},
		},
		want: testutil.MustDecodeHash("7b796833d01b092861681b937bde9b8a2c0c08721d4e12a9609e5827a4fffc54"),
	}, {
		witnesses: [][][]byte{
			{
				{1},
				[]byte("000000"),
			},
			{
				{1},
				[]byte("111111"),
			},
		},
		want: testutil.MustDecodeHash("d7e45ef3b47acaa13599d69a66896eed0bb713be4487efb512b0cd5bc5e2c349"),
	}, {
		witnesses: [][][]byte{
			{
				{1},
				[]byte("000000"),
			},
			{
				{2},
				[]byte("111111"),
				[]byte("222222"),
			},
		},
		want: testutil.MustDecodeHash("d7e45ef3b47acaa13599d69a66896eed0bb713be4487efb512b0cd5bc5e2c349"),
	}}

	for _, c := range cases {
		var txs []*bc.Tx
		for _, wit := range c.witnesses {
			txs = append(txs, NewTx(TxData{
				Inputs: []*TxInput{
					&TxInput{
						AssetVersion: 1,
						TypedInput: &SpendInput{
							Arguments: wit,
							SpendCommitment: SpendCommitment{
								AssetAmount: bc.AssetAmount{
									AssetId: &bc.AssetID{V0: 0},
								},
							},
						},
					},
				},
			}).Tx)
		}
		got, err := TxMerkleRoot(txs)
		if err != nil {
			t.Fatalf("unexpected error %s", err)
		}
		if got != c.want {
			t.Log("witnesses", c.witnesses)
			t.Errorf("got merkle root = %x want %x", got.Bytes(), c.want.Bytes())
		}
	}
}

func TestDuplicateLeaves(t *testing.T) {
	trueProg := []byte{byte(vm.OP_TRUE)}
	assetID := bc.ComputeAssetID(trueProg, 1, &bc.EmptyStringHash)
	txs := make([]*bc.Tx, 6)
	for i := uint64(0); i < 6; i++ {
		now := []byte(time.Now().String())
		txs[i] = NewTx(TxData{
			Version: 1,
			Inputs:  []*TxInput{NewIssuanceInput(now, i, trueProg, nil, nil)},
			Outputs: []*TxOutput{NewTxOutput(assetID, i, trueProg)},
		}).Tx
	}

	// first, get the root of an unbalanced tree
	txns := []*bc.Tx{txs[5], txs[4], txs[3], txs[2], txs[1], txs[0]}
	root1, err := TxMerkleRoot(txns)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	// now, get the root of a balanced tree that repeats leaves 0 and 1
	txns = []*bc.Tx{txs[5], txs[4], txs[3], txs[2], txs[1], txs[0], txs[1], txs[0]}
	root2, err := TxMerkleRoot(txns)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if root1 == root2 {
		t.Error("forged merkle tree by duplicating some leaves")
	}
}

func TestAllDuplicateLeaves(t *testing.T) {
	trueProg := []byte{byte(vm.OP_TRUE)}
	assetID := bc.ComputeAssetID(trueProg, 1, &bc.EmptyStringHash)
	now := []byte(time.Now().String())
	issuanceInp := NewIssuanceInput(now, 1, trueProg, nil, nil)

	tx := NewTx(TxData{
		Version: 1,
		Inputs:  []*TxInput{issuanceInp},
		Outputs: []*TxOutput{NewTxOutput(assetID, 1, trueProg)},
	}).Tx
	tx1, tx2, tx3, tx4, tx5, tx6 := tx, tx, tx, tx, tx, tx

	// first, get the root of an unbalanced tree
	txs := []*bc.Tx{tx6, tx5, tx4, tx3, tx2, tx1}
	root1, err := TxMerkleRoot(txs)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	// now, get the root of a balanced tree that repeats leaves 5 and 6
	txs = []*bc.Tx{tx6, tx5, tx6, tx5, tx4, tx3, tx2, tx1}
	root2, err := TxMerkleRoot(txs)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if root1 == root2 {
		t.Error("forged merkle tree with all duplicate leaves")
	}
}

func TestTxMerkleProof(t *testing.T) {
	cases := []struct {
		txCount          int
		relatedTxIndexes []int
		expectHashLen    int
		expectFlags      []uint8
	}{
		{
			txCount:          10,
			relatedTxIndexes: []int{0, 3, 7, 8},
			expectHashLen:    9,
			expectFlags:      []uint8{1, 1, 1, 1, 2, 0, 1, 0, 2, 1, 0, 1, 0, 2, 1, 2, 0},
		},
		{
			txCount:          10,
			relatedTxIndexes: []int{},
			expectHashLen:    1,
			expectFlags:      []uint8{0},
		},
		{
			txCount:          1,
			relatedTxIndexes: []int{0},
			expectHashLen:    1,
			expectFlags:      []uint8{2},
		},
		{
			txCount:          19,
			relatedTxIndexes: []int{1, 3, 5, 7, 11, 15},
			expectHashLen:    15,
			expectFlags:      []uint8{1, 1, 1, 1, 1, 0, 2, 1, 0, 2, 1, 1, 0, 2, 1, 0, 2, 1, 1, 0, 1, 0, 2, 1, 0, 1, 0, 2, 0},
		},
	}
	for _, c := range cases {
		txs, bcTxs := mockTransactions(c.txCount)

		var nodes []merkleNode
		for _, tx := range txs {
			nodes = append(nodes, tx.ID)
		}
		tree := buildMerkleTree(nodes)
		root, err := TxMerkleRoot(bcTxs)
		if err != nil {
			t.Fatalf("unexpected error %s", err)
		}
		if tree.hash != root {
			t.Error("build tree fail")
		}

		var relatedTx []*Tx
		for _, index := range c.relatedTxIndexes {
			relatedTx = append(relatedTx, txs[index])
		}
		proofHashes, flags := GetTxMerkleTreeProof(txs, relatedTx)
		if !testutil.DeepEqual(flags, c.expectFlags) {
			t.Error("The flags is not equals expect flags", flags, c.expectFlags)
		}
		if len(proofHashes) != c.expectHashLen {
			t.Error("The length proof hashes is not equals expect length")
		}
		var ids []*bc.Hash
		for _, tx := range relatedTx {
			ids = append(ids, &tx.ID)
		}
		if !ValidateTxMerkleTreeProof(proofHashes, flags, ids, root) {
			t.Error("Merkle tree validate fail")
		}
	}
}

func TestStatusMerkleProof(t *testing.T) {
	cases := []struct {
		statusCount    int
		relatedIndexes []int
		flags          []uint8
		expectHashLen  int
	}{
		{
			statusCount:    10,
			relatedIndexes: []int{0, 3, 7, 8},
			flags:          []uint8{1, 1, 1, 1, 2, 0, 1, 0, 2, 1, 0, 1, 0, 2, 1, 2, 0},
			expectHashLen:  9,
		},
		{
			statusCount:    10,
			relatedIndexes: []int{},
			flags:          []uint8{0},
			expectHashLen:  1,
		},
		{
			statusCount:    1,
			relatedIndexes: []int{0},
			flags:          []uint8{2},
			expectHashLen:  1,
		},
		{
			statusCount:    19,
			relatedIndexes: []int{1, 3, 5, 7, 11, 15},
			flags:          []uint8{1, 1, 1, 1, 1, 0, 2, 1, 0, 2, 1, 1, 0, 2, 1, 0, 2, 1, 1, 0, 1, 0, 2, 1, 0, 1, 0, 2, 0},
			expectHashLen:  15,
		},
	}
	for _, c := range cases {
		statuses := mockStatuses(c.statusCount)
		var relatedStatuses []*bc.TxVerifyResult
		for _, index := range c.relatedIndexes {
			relatedStatuses = append(relatedStatuses, statuses[index])
		}
		hashes := GetStatusMerkleTreeProof(statuses, c.flags)
		if len(hashes) != c.expectHashLen {
			t.Error("The length proof hashes is not equals expect length")
		}
		root, _ := TxStatusMerkleRoot(statuses)
		if !ValidateStatusMerkleTreeProof(hashes, c.flags, relatedStatuses, root) {
			t.Error("Merkle tree validate fail")
		}
	}
}

func convertHashStr2Bytes(hashStr string) ([32]byte, error) {
	var result [32]byte
	hashBytes, err := hex.DecodeString(hashStr)
	if err != nil {
		return result, err
	}
	copy(result[:], hashBytes)
	return result, nil
}

func mockTransactions(txCount int) ([]*Tx, []*bc.Tx) {
	var txs []*Tx
	var bcTxs []*bc.Tx
	trueProg := []byte{byte(vm.OP_TRUE)}
	assetID := bc.ComputeAssetID(trueProg, 1, &bc.EmptyStringHash)
	for i := 0; i < txCount; i++ {
		now := []byte(time.Now().String())
		issuanceInp := NewIssuanceInput(now, 1, trueProg, nil, nil)
		tx := NewTx(TxData{
			Version: 1,
			Inputs:  []*TxInput{issuanceInp},
			Outputs: []*TxOutput{NewTxOutput(assetID, 1, trueProg)},
		})
		txs = append(txs, tx)
		bcTxs = append(bcTxs, tx.Tx)
	}
	return txs, bcTxs
}

func mockStatuses(statusCount int) []*bc.TxVerifyResult {
	var statuses []*bc.TxVerifyResult
	for i := 0; i < statusCount; i++ {
		status := &bc.TxVerifyResult{}
		fail := rand.Intn(2)
		if fail == 0 {
			status.StatusFail = true
		} else {
			status.StatusFail = false
		}
		statuses = append(statuses, status)
	}
	return statuses
}
