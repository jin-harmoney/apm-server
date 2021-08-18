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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/beats/v7/libbeat/common"
)

func TestSpanTransform(t *testing.T) {
	path := "test/path"
	start := 0.65
	hexID, parentID, traceID := "0147258369012345", "abcdef0123456789", "01234567890123456789abcdefa"
	subtype := "amqp"
	action := "publish"
	timestamp := time.Date(2019, 1, 3, 15, 17, 4, 908.596*1e6,
		time.FixedZone("+0100", 3600))
	timestampUs := timestamp.UnixNano() / 1000
	method, statusCode, url := "get", 200, "http://localhost"
	instance, statement, dbType, user, rowsAffected := "db01", "select *", "sql", "jane", 5
	destServiceType, destServiceName, destServiceResource := "db", "elasticsearch", "elasticsearch"

	tests := []struct {
		Span   Span
		Output common.MapStr
		Msg    string
	}{
		{
			Msg:  "Span without a Stacktrace",
			Span: Span{},
			Output: common.MapStr{
				"processor": common.MapStr{"event": "span", "name": "transaction"},
				"span": common.MapStr{
					"duration": common.MapStr{"us": 0},
					"name":     "",
					"type":     "",
				},
				"timestamp": common.MapStr{"us": timestampUs},
				"url": common.MapStr{
					"original": url,
				},
			},
		},
		{
			Msg: "Full Span",
			Span: Span{
				ID:                  hexID,
				TraceID:             traceID,
				ParentID:            parentID,
				Name:                "myspan",
				Type:                "myspantype",
				Subtype:             subtype,
				Action:              action,
				Start:               &start,
				RepresentativeCount: 5,
				Duration:            1.20,
				Stacktrace:          Stacktrace{{AbsPath: path}},
				HTTP: &HTTP{
					Request:  &HTTPRequest{Method: method},
					Response: &HTTPResponse{StatusCode: statusCode},
				},
				DB: &DB{
					Instance:     instance,
					Statement:    statement,
					Type:         dbType,
					UserName:     user,
					RowsAffected: &rowsAffected,
				},
				DestinationService: &DestinationService{
					Type:     destServiceType,
					Name:     destServiceName,
					Resource: destServiceResource,
				},
				Message:   &Message{QueueName: "users"},
				Composite: &Composite{Count: 10, Sum: 1.1, CompressionStrategy: "exact_match"},
			},
			Output: common.MapStr{
				"span": common.MapStr{
					"id":       hexID,
					"duration": common.MapStr{"us": 1200},
					"name":     "myspan",
					"start":    common.MapStr{"us": 650},
					"type":     "myspantype",
					"subtype":  subtype,
					"action":   action,
					"stacktrace": []common.MapStr{{
						"exclude_from_grouping": false,
						"abs_path":              path,
					}},
					"db": common.MapStr{
						"instance":      instance,
						"statement":     statement,
						"type":          dbType,
						"user":          common.MapStr{"name": user},
						"rows_affected": rowsAffected,
					},
					"http": common.MapStr{
						"response": common.MapStr{"status_code": statusCode},
						"method":   "get",
					},
					"http.url.original": url,
					"destination": common.MapStr{
						"service": common.MapStr{
							"type":     destServiceType,
							"name":     destServiceName,
							"resource": destServiceResource,
						},
					},
					"message": common.MapStr{"queue": common.MapStr{"name": "users"}},
					"composite": common.MapStr{
						"count":                10,
						"sum":                  common.MapStr{"us": 1100},
						"compression_strategy": "exact_match",
					},
				},
				"processor": common.MapStr{"event": "span", "name": "transaction"},
				"timestamp": common.MapStr{"us": timestampUs},
				"trace":     common.MapStr{"id": traceID},
				"parent":    common.MapStr{"id": parentID},
				"http": common.MapStr{
					"response": common.MapStr{"status_code": statusCode},
					"request":  common.MapStr{"method": "get"},
				},
				"url": common.MapStr{
					"original": url,
				},
			},
		},
	}

	for _, test := range tests {
		event := APMEvent{
			Span:      &test.Span,
			Timestamp: timestamp,
			URL:       URL{Original: url},
		}
		output := event.BeatEvent(context.Background())
		assert.Equal(t, test.Output, output.Fields, test.Msg)
	}
}
