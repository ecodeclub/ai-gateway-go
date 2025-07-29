// Copyright 2025 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sn

import (
	"strconv"
	"sync/atomic"
	"time"
)

const (
	// 基准时间 - 2024年1月1日，可以根据实际需求调整
	epochMillis = int64(1704067200000) // 2024-01-01 00:00:00 UTC in milliseconds
	// 位数分配常量
	timestampBits = 41 // 时间戳位数
	hashBits      = 10 // hash值位数
	sequenceBits  = 12 // 序列号位数

	// 位移常量
	sequenceShift  = 0 // 别删了
	hashShift      = sequenceBits
	timestampShift = hashBits + sequenceBits

	// 掩码常量
	sequenceMask  = (1 << sequenceBits) - 1
	hashMask      = (1 << hashBits) - 1
	timestampMask = (1 << timestampBits) - 1
)

type Generator struct {
	// 目前暂时固定就是使用
	sequence int64
}

// Generate SN 有极小的概率冲突，也就是一个用户同时操作
// 并且还被路由到了不同机器上。所以这里我们懒得处理了
// 毕竟这种用户不是一个正常的用户。
func (g *Generator) Generate(uid int64) string {
	timestamp := time.Now().UnixMilli() - epochMillis
	// 计算hash值并取余
	// 使用原子操作安全地递增序列号
	sequence := atomic.AddInt64(&g.sequence, 1)

	// 组装最终ID
	id := (timestamp&timestampMask)<<timestampShift | // 时间戳部分
		(uid&hashMask)<<hashShift | // hash值部分
		(sequence & sequenceMask) // 序列号部分
	// 将 id 转为 36 进制编码
	return strconv.FormatInt(id, 36)
}
