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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/browser"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

var (
	errAddonDocsNotFound    = errors.New("addon not found")
	errAddonDocsUnavailable = errors.New("addon has no documentation")
	openAddonDocsURL        = browser.OpenURL
)

var addonsDocsCmd = &cobra.Command{
	Use:     "docs ADDON_NAME",
	Short:   "Open documentation for ADDON_NAME in a browser",
	Long:    "Open documentation for ADDON_NAME in a browser. For a list of available addons use: minikube addons list",
	Example: "minikube addons docs dashboard",
	Run: func(_ *cobra.Command, args []string) {
		if len(args) != 1 {
			exit.Message(reason.Usage, "usage: minikube addons docs ADDON_NAME")
		}

		addonName := args[0]
		docsURL, err := addonDocsURL(addonName)
		if errors.Is(err, errAddonDocsNotFound) {
			exit.Message(reason.Usage, `addon '{{.name}}' is not a valid addon packaged with minikube.
To see the list of available addons run:
minikube addons list`, out.V{"name": addonName})
		}
		if errors.Is(err, errAddonDocsUnavailable) {
			out.Infof(`{{.name}} doesn't have documentation.
To see docs links for all addons run:
minikube addons list --docs`, out.V{"name": addonName})
			return
		}
		if err != nil {
			exit.Error(reason.InternalAddonEnable, "failed to resolve addon documentation", err)
		}

		out.Styled(style.Celebrate, "Opening {{.url}} in your default browser...", out.V{"url": docsURL})
		if err := openAddonDocsURL(docsURL); err != nil {
			exit.Error(reason.HostBrowser, fmt.Sprintf("browser failed to open url %s", docsURL), err)
		}
	},
}

func addonDocsURL(addonName string) (string, error) {
	addon, ok := assets.Addons[addonName]
	if !ok {
		return "", errAddonDocsNotFound
	}

	docsURL := strings.TrimSpace(addon.Docs)
	if docsURL == "" {
		return "", errAddonDocsUnavailable
	}
	return docsURL, nil
}

func init() {
	AddonsCmd.AddCommand(addonsDocsCmd)
}
