// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package config

import (
	"strings"

	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/kibana"
)

type KibanaConfig struct {
	Enabled             bool   `config:"enabled"`
	APIKey              string `config:"api_key"`
	kibana.ClientConfig `config:",inline"`
}

func (k *KibanaConfig) Unpack(cfg *config.C) error {
	type kibanaConfig KibanaConfig
	if err := cfg.Unpack((*kibanaConfig)(k)); err != nil {
		return err
	}
	k.Enabled = cfg.Enabled()
	k.Host = strings.TrimRight(k.Host, "/")
	return nil
}

func defaultKibanaConfig() KibanaConfig {
	return KibanaConfig{
		Enabled:      false,
		ClientConfig: kibana.DefaultClientConfig(),
	}
}
