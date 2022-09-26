// Copyright 2022 The AmazeChain Authors
// This file is part of the AmazeChain library.
//
// The AmazeChain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The AmazeChain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the AmazeChain library. If not, see <http://www.gnu.org/licenses/>.

package rawdb

import (
	"fmt"
	"github.com/amazechain/amc/api/protocol/types_pb"
	"github.com/amazechain/amc/common/block"
	"github.com/amazechain/amc/common/db"
	"github.com/amazechain/amc/common/transaction"
	"github.com/amazechain/amc/common/types"
	"github.com/gogo/protobuf/proto"
)

func GetGenesis(db db.IDatabase) (block.IBlock, error) {
	header, _, err := GetHeader(db, types.NewInt64(0))
	if err != nil {
		return nil, err
	}
	body, err := GetBody(db, types.NewInt64(0))

	if err != nil {
		return nil, err
	}

	genesisBlock := block.NewBlock(header, body.Transactions())

	return genesisBlock, nil
}

func StoreGenesis(db db.IDatabase, genesisBlock block.IBlock) error {
	return StoreBlock(db, genesisBlock)
}

func GetHeader(db db.IDatabase, number types.Int256) (block.IHeader, types.Hash, error) {
	key, err := number.MarshalText()
	if err != nil {
		return nil, types.Hash{}, err
	}

	r, err := db.OpenReader(headerDB)
	if err != nil {
		return nil, types.Hash{}, err
	}

	v, err := r.Get(key)
	if err != nil {
		return nil, types.Hash{}, err
	}

	var (
		header   block.Header
		pbHeader types_pb.PBHeader
	)

	if err := proto.Unmarshal(v, &pbHeader); err != nil {
		return nil, types.Hash{}, err
	}

	if err := header.FromProtoMessage(&pbHeader); err != nil {
		return nil, types.Hash{}, err
	}

	return &header, header.Hash(), nil
}

func StoreHeader(db db.IDatabase, iheader block.IHeader) error {
	header, ok := iheader.(*block.Header)
	if !ok {
		return fmt.Errorf("failed header types")
	}

	key, err := header.Number.MarshalText()
	if err != nil {
		return err
	}

	v, err := header.Marshal()
	if err != nil {
		return err
	}

	h := header.Hash()

	w, err := db.OpenWriter(headerDB)
	if err != nil {
		return err
	}

	_ = w.Put(key, v)
	_ = w.Put(h[:], key)

	return nil
}

func GetHeaderByHash(db db.IDatabase, hash types.Hash) (block.IHeader, types.Hash, error) {
	r, err := db.OpenReader(headerDB)
	if err != nil {
		return nil, types.Hash{}, err
	}

	key, err := r.Get(hash[:])
	if err != nil {
		return nil, types.Hash{}, err
	}

	v, err := r.Get(key)
	if err != nil {
		return nil, types.Hash{}, err
	}

	var (
		header   block.Header
		pbHeader types_pb.PBHeader
	)

	if err := proto.Unmarshal(v, &pbHeader); err != nil {
		return nil, types.Hash{}, err
	}

	if err := header.FromProtoMessage(&pbHeader); err != nil {
		return nil, types.Hash{}, err
	}

	return &header, header.Hash(), nil
}

func SaveBlocks(db db.IDatabase, blocks []block.IBlock) (int, error) {

	// todo  batch?
	for i, block := range blocks {
		if err := StoreBlock(db, block); err != nil {
			return i + 1, err
		}
	}

	return 0, nil
}

func StoreBlock(db db.IDatabase, block block.IBlock) error {

	if err := StoreHeader(db, block.Header()); err != nil {
		return err
	}

	if err := StoreBody(db, block.Number64(), block.Body()); err != nil {
		return err
	}

	if err := StoreHashNumber(db, block.Header().Hash(), block.Number64()); err != nil {
		return err
	}
	// index
	for _, tx := range block.Transactions() {
		hash, _ := tx.Hash()
		if err := StoreTransactionIndex(db, block.Number64(), hash); err != nil {
			return err
		}
	}
	return nil
}

func GetBody(db db.IDatabase, number types.Int256) (block.IBody, error) {
	key, err := number.MarshalText()
	if err != nil {
		return nil, err
	}

	r, err := db.OpenReader(bodyDB)
	if err != nil {
		return nil, err
	}

	v, err := r.Get(key)
	if err != nil {
		return nil, err
	}

	var (
		pBody types_pb.PBody
		body  block.Body
	)
	if err := proto.Unmarshal(v, &pBody); err != nil {
		return nil, err
	}

	err = body.FromProtoMessage(&pBody)

	return &body, err
}

func StoreBody(db db.IDatabase, number types.Int256, body block.IBody) error {
	key, err := number.MarshalText()
	if err != nil {
		return err
	}
	if body == nil {
		return nil
	}

	pbBody := body.ToProtoMessage()

	v, err := proto.Marshal(pbBody)
	if err != nil {
		return err
	}

	w, err := db.OpenWriter(bodyDB)
	if err != nil {
		return err
	}

	return w.Put(key, v)
}

func GetLatestBlock(db db.IDatabase) (block.IBlock, error) {
	key := types.StringToHash(latestBlock)
	r, err := db.OpenReader(latestStateDB)
	if err != nil {
		return nil, err
	}
	v, err := r.Get(key[:])
	if err != nil {
		return nil, err
	}
	var (
		pBlock types_pb.PBlock
		block  block.Block
	)
	if err := proto.Unmarshal(v, &pBlock); err != nil {
		return nil, err
	}
	if err := block.FromProtoMessage(&pBlock); err != nil {
		return nil, err
	}

	return &block, nil
}

func SaveLatestBlock(db db.IDatabase, block block.IBlock) error {
	key := types.StringToHash(latestBlock)
	b, err := proto.Marshal(block.ToProtoMessage())
	if err != nil {
		return err
	}

	w, err := db.OpenWriter(latestStateDB)
	if err != nil {
		return err
	}

	return w.Put(key[:], b)
}

// GetTransaction get tx by txHash
func GetTransaction(db db.IDatabase, txHash types.Hash) (*transaction.Transaction, types.Hash, types.Int256, uint64, error) {

	blockNumber, err := GetTransactionIndex(db, txHash)
	if err != nil {
		return nil, types.Hash{}, types.NewInt64(0), 0, nil
	}
	body, err := GetBody(db, blockNumber)
	if err != nil {
		return nil, types.Hash{}, types.NewInt64(0), 0, nil
	}
	_, headerHash, err := GetHeader(db, blockNumber)
	if err != nil {
		return nil, types.Hash{}, types.NewInt64(0), 0, nil
	}
	for txIndex, tx := range body.Transactions() {
		hash, _ := tx.Hash()
		if hash == txHash {
			return tx, headerHash, blockNumber, uint64(txIndex), nil
		}
	}

	return nil, types.Hash{}, types.NewInt64(0), 0, nil
}

// StoreTransactionIndex store txHash blockNumber index
func StoreTransactionIndex(db db.IDatabase, blockNumber types.Int256, txHash types.Hash) error {
	v, err := blockNumber.MarshalText()
	if err != nil {
		return err
	}
	r, err := db.OpenWriter(transactionIndex)
	if err != nil {
		return err
	}

	return r.Put(txHash.HexBytes(), v)
}

// GetTransactionIndex get block number by txHash
func GetTransactionIndex(db db.IDatabase, txHash types.Hash) (types.Int256, error) {

	r, err := db.OpenReader(transactionIndex)
	if err != nil {
		return types.NewInt64(0), err
	}

	v, err := r.Get(txHash.HexBytes())

	if err != nil {
		return types.NewInt64(0), err
	}

	int256, err := types.FromHex(string(v))

	if err != nil {
		return types.NewInt64(0), err
	}
	return int256, nil
}

// StoreHashNumber store hash to number index
func StoreHashNumber(db db.IDatabase, hash types.Hash, number types.Int256) error {

	v, err := number.MarshalText()
	if err != nil {
		return err
	}
	r, err := db.OpenWriter(hashNumberDB)
	if err != nil {
		return err
	}

	return r.Put(hash.HexBytes(), v)
}

// GetHashNumber get hash to number index
func GetHashNumber(db db.IDatabase, hash types.Hash) (types.Int256, error) {

	r, err := db.OpenReader(hashNumberDB)
	if err != nil {
		return types.NewInt64(0), err
	}

	v, err := r.Get(hash.HexBytes())

	if err != nil {
		return types.NewInt64(0), err
	}

	int256, err := types.FromHex(string(v))

	if err != nil {
		return types.NewInt64(0), err
	}
	return int256, nil
}

func GetReceipts(db db.IDatabase, hash types.Hash) (block.Receipts, error) {

	r, err := db.OpenReader(receiptsDB)
	if err != nil {
		return nil, err
	}

	v, err := r.Get(hash.HexBytes())
	if err != nil {
		return nil, err
	}

	var (
		pReceipts types_pb.Receipts
		receipts  block.Receipts
	)
	if err := proto.Unmarshal(v, &pReceipts); err != nil {
		return nil, err
	}

	err = receipts.FromProtoMessage(&pReceipts)

	return receipts, err
}

func StoreReceipts(db db.IDatabase, hash types.Hash, receipts block.Receipts) error {
	if receipts == nil {
		return nil
	}

	pReceipts := receipts.ToProtoMessage()

	v, err := proto.Marshal(pReceipts)
	if err != nil {
		return err
	}

	w, err := db.OpenWriter(receiptsDB)
	if err != nil {
		return err
	}

	return w.Put(hash.HexBytes(), v)
}

func ReadTd(db db.IDatabase, hash types.Hash) types.Int256 {
	return types.Int256{}
}

func WriteTD(db db.IDatabase, hash types.Hash, td types.Int256) error {
	return nil
}

func DeleteTD(db db.IDatabase, hash types.Hash) {
}