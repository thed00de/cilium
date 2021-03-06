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
package types

import (
	"fmt"

	"github.com/cilium/cilium/common"
)

type StatusCode int

const (
	OK       StatusCode = 0
	Warning  StatusCode = -1
	Failure  StatusCode = -2
	Disabled StatusCode = -3
)

func NewStatusOK(info string) Status {
	if info == "" {
		return Status{Code: OK, Msg: "OK"}
	} else {
		return Status{Code: OK, Msg: info}
	}
}

type Status struct {
	Code StatusCode `json:"code"`
	Msg  string     `json:"msg"`
}

func (s Status) String() string {
	var text string
	switch s.Code {
	case OK:
		text = common.Green("OK")
		return fmt.Sprintf("%s", text)
	case Warning:
		text = common.Yellow("Warning")
		return fmt.Sprintf("%s - %s", text, s.Msg)
	case Failure:
		text = common.Red("Failure")
		return fmt.Sprintf("%s - %s", text, s.Msg)
	case Disabled:
		text = common.Yellow("Disabled")
		return fmt.Sprintf("%s - %s", text, s.Msg)
	}
	return "Unknown code"
}

type StatusResponse struct {
	Consul     Status `json:"consul"`
	Docker     Status `json:"docker"`
	Kubernetes Status `json:"kubernetes"`
	Logstash   Status `json:"logstash"`
	Cilium     Status `json:"cilium"`
}
