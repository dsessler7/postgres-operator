// Copyright 2024 - 2025 Crunchy Data Solutions, Inc.
//
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/crunchydata/postgres-operator/pkg/apis/postgres-operator.crunchydata.com/v1beta1"
)

func TestConfigToYAML(t *testing.T) {
	t.Run("NilInstrumentationSpec", func(t *testing.T) {
		result, err := NewConfig(nil).ToYAML()
		assert.NilError(t, err)
		assert.DeepEqual(t, result, `# Generated by postgres-operator. DO NOT EDIT.
# Your changes will not be saved.
exporters:
  debug:
    verbosity: detailed
extensions: {}
processors:
  batch/1s:
    timeout: 1s
  batch/200ms:
    timeout: 200ms
  groupbyattrs/compact: {}
receivers: {}
service:
  extensions: []
  pipelines: {}
`)
	})

	t.Run("InstrumentationSpecDefined", func(t *testing.T) {
		spec := testInstrumentationSpec()

		result, err := NewConfig(spec).ToYAML()
		assert.NilError(t, err)
		assert.DeepEqual(t, result, `# Generated by postgres-operator. DO NOT EDIT.
# Your changes will not be saved.
exporters:
  debug:
    verbosity: detailed
  googlecloud:
    log:
      default_log_name: opentelemetry.io/collector-exported-log
    project: google-project-name
extensions: {}
processors:
  batch/1s:
    timeout: 1s
  batch/200ms:
    timeout: 200ms
  groupbyattrs/compact: {}
receivers: {}
service:
  extensions: []
  pipelines: {}
`)
	})
}

// TODO: write this test after rebasing on new retention API changes.
func TestGenerateLogrotateConfig(t *testing.T) {
	for _, tt := range []struct {
		logFilePath      string
		retentionPeriod  string
		postrotateScript string
		result           string
	}{
		{
			logFilePath:      "/this/is/a/file.path",
			retentionPeriod:  "12h",
			postrotateScript: "echo 'Hello, World'",
			result: `/this/is/a/file.path {
      rotate 12
      missingok
      sharedscripts
      notifempty
      nocompress
      hourly
      postrotate
            echo 'Hello, World'
      endscript
}
`,
		},
		{
			logFilePath:      "/tmp/test.log",
			retentionPeriod:  "5 days",
			postrotateScript: "",
			result: `/tmp/test.log {
      rotate 5
      missingok
      sharedscripts
      notifempty
      nocompress
      daily
      postrotate
            
      endscript
}
`,
		},
		{
			logFilePath:      "/tmp/test.log",
			retentionPeriod:  "5wk",
			postrotateScript: "pkill -HUP --exact pgbouncer",
			result: `/tmp/test.log {
      rotate 35
      missingok
      sharedscripts
      notifempty
      nocompress
      daily
      postrotate
            pkill -HUP --exact pgbouncer
      endscript
}
`,
		},
	} {
		t.Run(tt.retentionPeriod, func(t *testing.T) {
			duration, err := v1beta1.NewDuration(tt.retentionPeriod)
			assert.NilError(t, err)
			result := generateLogrotateConfig(tt.logFilePath, duration, tt.postrotateScript)
			assert.Equal(t, tt.result, result)
		})
	}
}

func TestParseDurationForLogrotate(t *testing.T) {
	for _, tt := range []struct {
		retentionPeriod string
		number          int
		interval        string
	}{
		{
			retentionPeriod: "12h",
			number:          12,
			interval:        "hourly",
		},
		{
			retentionPeriod: "24hr",
			number:          1,
			interval:        "daily",
		},
		{
			retentionPeriod: "35hour",
			number:          1,
			interval:        "daily",
		},
		{
			retentionPeriod: "36 hours",
			number:          2,
			interval:        "daily",
		},
		{
			retentionPeriod: "3d",
			number:          3,
			interval:        "daily",
		},
		{
			retentionPeriod: "365day",
			number:          365,
			interval:        "daily",
		},
		{
			retentionPeriod: "1w",
			number:          7,
			interval:        "daily",
		},
		{
			retentionPeriod: "4wk",
			number:          28,
			interval:        "daily",
		},
		{
			retentionPeriod: "52week",
			number:          364,
			interval:        "daily",
		},
	} {
		t.Run(tt.retentionPeriod, func(t *testing.T) {
			duration, err := v1beta1.NewDuration(tt.retentionPeriod)
			assert.NilError(t, err)
			number, interval := parseDurationForLogrotate(duration)
			assert.Equal(t, tt.number, number)
			assert.Equal(t, tt.interval, interval)
		})
	}
}
