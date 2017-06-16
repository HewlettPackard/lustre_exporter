// (C) Copyright 2017 Hewlett Packard Enterprise Development LP
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

package sources

import (
	"github.com/prometheus/client_golang/prometheus"
)

// ProcLocation is the source to pull proc files from. By default, use the '/proc' directory on the local node,
// but for testing purposes, specify 'proc' (without the leading '/') for the local files.
var ProcLocation = "/proc"

//Namespace defines the namespace shared by all Lustre metrics.
const Namespace = "lustre"

//Factories contains the list of all sources.
var Factories = make(map[string]func() LustreSource)

//LustreSource is the interface that each source implements.
type LustreSource interface {
	Update(ch chan<- prometheus.Metric) (err error)
}
