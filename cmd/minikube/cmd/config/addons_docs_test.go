/*
Copyright 2026 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/out"
)

func captureAddonsDocsStdout(t *testing.T, f func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer r.Close()
	old := os.Stdout
	defer func() {
		os.Stdout = old
		out.SetOutFile(old)
	}()
	os.Stdout = w
	out.SetOutFile(w)

	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()

	f()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe: %v", err)
	}

	return <-done
}

func TestAddonDocsURL(t *testing.T) {
	tests := []struct {
		name    string
		addon   string
		want    string
		wantErr error
	}{
		{
			name:  "handbook docs",
			addon: "dashboard",
			want:  "https://minikube.sigs.k8s.io/docs/handbook/dashboard/",
		},
		{
			name:  "handbook addon docs",
			addon: "ambassador",
			want:  "https://minikube.sigs.k8s.io/docs/handbook/addons/ambassador/",
		},
		{
			name:  "external docs",
			addon: "freshpod",
			want:  "https://github.com/GoogleCloudPlatform/freshpod",
		},
		{
			name:  "non-handbook minikube docs",
			addon: "nvidia-device-plugin",
			want:  "https://minikube.sigs.k8s.io/docs/tutorials/nvidia/",
		},
		{
			name:    "addon without docs",
			addon:   "auto-pause",
			wantErr: errAddonDocsUnavailable,
		},
		{
			name:    "unknown addon",
			addon:   "not-an-addon",
			wantErr: errAddonDocsNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addonDocsURL(tt.addon)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got != tt.want {
				t.Errorf("expected docs URL %q, got %q", tt.want, got)
			}
		})
	}
}

func TestAddonsDocsCmdOpensDocs(t *testing.T) {
	originalOpenAddonDocsURL := openAddonDocsURL
	defer func() {
		openAddonDocsURL = originalOpenAddonDocsURL
	}()

	var got string
	openAddonDocsURL = func(url string) error {
		got = url
		return nil
	}

	addonsDocsCmd.Run(addonsDocsCmd, []string{"dashboard"})

	want := "https://minikube.sigs.k8s.io/docs/handbook/dashboard/"
	if got != want {
		t.Errorf("expected browser to open %q, got %q", want, got)
	}
}

func TestAddonsDocsCmdWithNoDocsDoesNotOpenBrowser(t *testing.T) {
	originalOpenAddonDocsURL := openAddonDocsURL
	defer func() {
		openAddonDocsURL = originalOpenAddonDocsURL
	}()

	var opened bool
	openAddonDocsURL = func(url string) error {
		opened = true
		return nil
	}

	s := captureAddonsDocsStdout(t, func() {
		addonsDocsCmd.Run(addonsDocsCmd, []string{"volcano"})
	})

	if opened {
		t.Errorf("expected browser to not open for addon without docs")
	}
	if !strings.Contains(s, "volcano doesn't have documentation") {
		t.Errorf("expected no-docs message, got: %q", s)
	}
	if strings.Contains(s, "Exiting due to MK_USAGE") {
		t.Errorf("expected no-docs message to be non-fatal, got: %q", s)
	}
}

func TestAddonsHelpIncludesDocsCommand(t *testing.T) {
	if cmd, _, err := AddonsCmd.Find([]string{"docs"}); err != nil || cmd != addonsDocsCmd {
		t.Fatalf("expected addons docs command to be registered, got cmd=%v err=%v", cmd, err)
	}

	usage := AddonsCmd.UsageString()
	if !strings.Contains(usage, "docs") || !strings.Contains(usage, "ADDON_NAME") {
		t.Errorf("expected addons help to include docs usage, got: %q", usage)
	}
}
