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

package common

import (
	"github.com/amazechain/amc/common/block"
	"github.com/amazechain/amc/common/types"
	"github.com/amazechain/amc/internal/consensus"
	"github.com/libp2p/go-libp2p-core/peer"
)

type IHeaderChain interface {
	GetHeaderByNumber(number types.Int256) (block.IHeader, error)
	GetHeaderByHash(h types.Hash) (block.IHeader, error)
	InsertHeader(headers []block.IHeader) (int, error)
	GetBlockByHash(h types.Hash) (block.IBlock, error)
	GetBlockByNumber(number types.Int256) (block.IBlock, error)
}

type IBlockChain interface {
	IHeaderChain
	CurrentBlock() block.IBlock
	Blocks() []block.IBlock
	Start() error
	GenesisBlock() block.IBlock
	NewBlockHandler(payload []byte, peer peer.ID) error
	InsertChain(blocks []block.IBlock) (int, error)
	InsertBlock(blocks []block.IBlock, isSync bool) (int, error)
	SetEngine(engine consensus.Engine)
	GetBlocksFromHash(hash types.Hash, n int) (blocks []block.IBlock)
	SealedBlock(b block.IBlock)
	Engine() consensus.Engine
	GetReceipts(blockHash types.Hash) (block.Receipts, error)
	GetLogs(blockHash types.Hash) ([][]*block.Log, error)
	SetHead(head uint64) error

	//
	GetHeader(types.Hash, types.Int256) block.IHeader
	// alias for GetBlocksFromHash?
	GetBlock(hash types.Hash) block.IBlock
	StateAt(root types.Hash) IStateDB

	GetTd(hash types.Hash) types.Int256
}