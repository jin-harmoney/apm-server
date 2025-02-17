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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/elastic-agent-libs/mapstr"
)

func TestObserverFields(t *testing.T) {
	tests := []struct {
		Observer Observer
		Fields   mapstr.M
	}{
		{
			Observer: Observer{},
			Fields:   nil,
		},
		{
			Observer: Observer{
				EphemeralID: "observer_ephemeral_id",
				Hostname:    "observer_hostname",
				ID:          "observer_id",
				Name:        "observer_name",
				Type:        "observer_type",
				Version:     "observer_version",
			},
			Fields: mapstr.M{
				"ephemeral_id": "observer_ephemeral_id",
				"hostname":     "observer_hostname",
				"id":           "observer_id",
				"name":         "observer_name",
				"type":         "observer_type",
				"version":      "observer_version",
			},
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.Fields, test.Observer.Fields())
	}
}
