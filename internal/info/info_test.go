// Copyright 2019 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package info

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/testutil"
)

func TestLoadManifestFromReceiptOrIndex(t *testing.T) {
	const pluginName = "some-plugin"
	plugin := testutil.NewPlugin().WithName(pluginName).
		WithPlatforms(testutil.NewPlatform().V()).V()

	yamlBytes, err := yaml.Marshal(plugin)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		prepare   func(environment.Paths, *testutil.TempDir)
		shouldErr bool
	}{
		{
			name: "manifest in receipts",
			prepare: func(paths environment.Paths, tmpDir *testutil.TempDir) {
				path := paths.PluginInstallReceiptPath(pluginName)
				tmpDir.Write(path, yamlBytes)
			},
		},
		{
			name: "manifest in index",
			prepare: func(paths environment.Paths, tmpDir *testutil.TempDir) {
				path := filepath.Join(paths.IndexPluginsPath(), pluginName+".yaml")
				tmpDir.Write(path, yamlBytes)
			},
		},
		{
			name: "invalid manifest in receipts",
			prepare: func(paths environment.Paths, tmpDir *testutil.TempDir) {
				path := paths.PluginInstallReceiptPath(pluginName)
				tmpDir.Write(path, []byte("invalid yaml file"))
			},
			shouldErr: true,
		},
		{
			name: "invalid manifest in index",
			prepare: func(paths environment.Paths, tmpDir *testutil.TempDir) {
				path := filepath.Join(paths.IndexPluginsPath(), pluginName+".yaml")
				tmpDir.Write(path, []byte("invalid yaml file"))
			},
			shouldErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.NewTempDir(t)
			defer cleanup()

			paths := environment.NewPaths(tmpDir.Root())
			test.prepare(paths, tmpDir)
			actual, err := LoadManifestFromReceiptOrIndex(paths, pluginName)

			if test.shouldErr {
				if err == nil {
					t.Error("LoadManifestFromReceiptOrIndex expected an error but found none")
				}
			} else {
				if err != nil {
					t.Error(err)
					return
				}
				if diff := cmp.Diff(&plugin, &actual); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

func TestLoadManifestFromReceiptOrIndex_ReturnsIsNotExist(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	paths := environment.NewPaths(tmpDir.Root())
	_, err := LoadManifestFromReceiptOrIndex(paths, "non-existing-plugin")

	if err == nil {
		t.Fatalf("Expected LoadManifestFromReceiptOrIndex to fail")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("Expected error to be ENOENT but was %q", err)
	}
}
