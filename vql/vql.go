/*
   Velociraptor - Hunting Evil
   Copyright (C) 2019 Velocidex Innovations.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
/*

  The VQL subsystem allows for collecting host based state information
  using Velocidex Query Language (VQL) queries.

  The primary use case for Velociraptor is for incident
  response/detection and host based inventory management.
*/

package vql

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"www.velocidex.com/golang/vfilter"
)

var (
	exportedPlugins      = make(map[string]vfilter.PluginGeneratorInterface)
	exportedProtocolImpl []vfilter.Any
	exportedFunctions    = make(map[string]vfilter.FunctionInterface)

	scopeCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vql_make_scope",
		Help: "Total number of Scope objects constructed.",
	})
)

func RegisterPlugin(plugin vfilter.PluginGeneratorInterface) {
	name := plugin.Info(nil, nil).Name
	_, pres := exportedPlugins[name]
	if pres {
		panic("Multiple plugins defined")
	}

	exportedPlugins[name] = plugin
}

func RegisterFunction(plugin vfilter.FunctionInterface) {
	name := plugin.Info(nil, nil).Name
	_, pres := exportedFunctions[name]
	if pres {
		panic("Multiple plugins defined")
	}

	exportedFunctions[name] = plugin
}

func RegisterProtocol(plugin vfilter.Any) {
	exportedProtocolImpl = append(exportedProtocolImpl, plugin)
}

var (
	mu sync.Mutex

	// Instead of building the scope from scratch each time, use a
	// global scope and prepare any other scopes from it.
	globalScope *vfilter.Scope
)

func _makeRootScope() *vfilter.Scope {
	mu.Lock()
	defer mu.Unlock()

	if globalScope == nil {
		globalScope = vfilter.NewScope()
		for _, plugin := range exportedPlugins {
			globalScope.AppendPlugins(plugin)
		}

		for _, protocol := range exportedProtocolImpl {
			globalScope.AddProtocolImpl(protocol)
		}

		for _, function := range exportedFunctions {
			globalScope.AppendFunctions(function)
		}
	}

	return globalScope.NewScope()
}

func MakeScope() *vfilter.Scope {
	scopeCounter.Inc()

	return _makeRootScope()
}
