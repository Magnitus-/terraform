// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"github.com/opentffoundation/opentf/version"
)

// Deprecated: Providers should use schema.Provider.TerraformVersion instead
func VersionString() string {
	return version.String()
}
