// Copyright 2021 ecodeclub
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

package dao

type TempQuota struct {
	ID        int64  `gorm:"primaryKey;autoIncrement;column:id"`
	UID       string `gorm:"column:uid"`
	Amount    int64  `gorm:"column:amount"`
	StartTime int64  `gorm:"column:start_time"`
	EndTime   int64  `gorm:"column:end_time"`
	Ctime     int64  `gorm:"column:ctime"`
	Utime     int64  `gorm:"column:utime"`
}

type Quota struct {
	ID     int64  `gorm:"primaryKey;autoIncrement;colum:id"`
	UID    string `gorm:"column:uid"`
	Amount int64  `gorm:"colum:amount"`
	Ctime  int64  `gorm:"column:ctime"`
	Utime  int64  `gorm:"column:utime"`
}
