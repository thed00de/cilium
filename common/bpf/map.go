//
// Copyright 2016 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package bpf

/*
#cgo CFLAGS: -I../../bpf/include
#include <linux/unistd.h>
#include <linux/bpf.h>
#include <sys/resource.h>
*/

import "C"

import (
	"unsafe"

	"bufio"
	"fmt"
	"os"
)

type MapType int

// This enumeration must be in sync with <linux/bpf.h>
const (
	MapTypeUnspec MapType = iota
	MapTypeHash
	MapTypeArray
	MapTypeProgArray
	MapTypePerfEventArray
	MapTypePerCPUHash
	MapTypePerCPUArray
	MapTypeStackTrace
	MapTypeCgroupArray
)

func (t MapType) String() string {
	switch t {
	case MapTypeHash:
		return "Hash"
	case MapTypeArray:
		return "Array"
	case MapTypeProgArray:
		return "Program array"
	case MapTypePerfEventArray:
		return "Event array"
	case MapTypePerCPUHash:
		return "Per-CPU hash"
	case MapTypePerCPUArray:
		return "Per-CPU array"
	case MapTypeStackTrace:
		return "Stack trace"
	case MapTypeCgroupArray:
		return "Cgroup array"
	}

	return "Unknown"
}

type MapObj interface {
	GetPtr() unsafe.Pointer
}

type MapInfo struct {
	MapType    MapType
	KeySize    uint32
	ValueSize  uint32
	MaxEntries uint32
	Flags      uint32
}

type Map struct {
	MapInfo
	fd     int
	path   string
	isOpen bool
}

func NewMap(path string, mapType MapType, keySize int, valueSize int, maxEntries int) *Map {
	return &Map{
		MapInfo: MapInfo{
			MapType:    mapType,
			KeySize:    uint32(keySize),
			ValueSize:  uint32(valueSize),
			MaxEntries: uint32(maxEntries),
		},
		path:   path,
		isOpen: false,
	}
}

func GetMapInfo(pid int, fd int) (*MapInfo, error) {
	fdinfoFile := fmt.Sprintf("/proc/%d/fdinfo/%d", pid, fd)

	file, err := os.Open(fdinfoFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info := &MapInfo{}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var value int

		line := scanner.Text()
		if n, err := fmt.Sscanf(line, "map_type:\t%d", &value); n == 1 && err == nil {
			info.MapType = MapType(value)
		} else if n, err := fmt.Sscanf(line, "key_size:\t%d", &value); n == 1 && err == nil {
			info.KeySize = uint32(value)
		} else if n, err := fmt.Sscanf(line, "value_size:\t%d", &value); n == 1 && err == nil {
			info.ValueSize = uint32(value)
		} else if n, err := fmt.Sscanf(line, "max_entries:\t%d", &value); n == 1 && err == nil {
			info.MaxEntries = uint32(value)
		} else if n, err := fmt.Sscanf(line, "map_flas:\t%i", &value); n == 1 && err == nil {
			info.Flags = uint32(value)
		}
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return info, nil
}

func OpenMap(path string) (*Map, error) {
	fd, err := ObjGet(path)
	if err != nil {
		return nil, err
	}

	info, err := GetMapInfo(os.Getpid(), fd)
	if err != nil {
		return nil, err
	}

	if info.MapType == 0 {
		return nil, fmt.Errorf("Unable to determine map type")
	}

	if info.KeySize == 0 {
		return nil, fmt.Errorf("Unable to determine map key size")
	}

	return &Map{
		MapInfo: *info,
		fd:      fd,
		path:    path,
		isOpen:  true,
	}, nil
}

func (m *Map) OpenOrCreate() (bool, error) {
	fd, isNew, err := OpenOrCreateMap(m.path, int(m.MapType), m.KeySize, m.ValueSize, m.MaxEntries)
	if err != nil {
		return false, err
	}

	m.fd = fd
	m.isOpen = true

	return isNew, nil
}

func (m *Map) Open() error {
	fd, err := ObjGet(m.path)
	if err != nil {
		return err
	}

	m.fd = fd
	m.isOpen = true

	return nil
}

type DumpFunc func(key []byte, value []byte)

func (m *Map) Dump(cb DumpFunc) error {
	key := make([]byte, m.KeySize)
	nextKey := make([]byte, m.KeySize)
	value := make([]byte, m.ValueSize)

	if !m.isOpen {
		if err := m.Open(); err != nil {
			return err
		}
	}

	for {
		err := GetNextKey(
			m.fd,
			unsafe.Pointer(&key[0]),
			unsafe.Pointer(&nextKey[0]),
		)

		if err != nil {
			break
		}

		err = LookupElement(
			m.fd,
			unsafe.Pointer(&nextKey[0]),
			unsafe.Pointer(&value[0]),
		)

		if err != nil {
			return err
		}

		cb(nextKey, value)
		copy(key, nextKey)
	}

	return nil
}

func (m *Map) Lookup(key MapObj, value unsafe.Pointer) error {
	if !m.isOpen {
		if err := m.Open(); err != nil {
			return err
		}
	}

	return LookupElement(m.fd, key.GetPtr(), value)
}

func (m *Map) Update(key MapObj, value unsafe.Pointer) error {
	if !m.isOpen {
		if err := m.Open(); err != nil {
			return err
		}
	}

	return UpdateElement(m.fd, key.GetPtr(), value, 0)
}

func (m *Map) Delete(key MapObj) error {
	if !m.isOpen {
		if err := m.Open(); err != nil {
			return err
		}
	}

	return DeleteElement(m.fd, key.GetPtr())
}
