// Copyright (c) 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"fmt"

	"github.com/vmware-tanzu/crash-diagnostics/script"
)

// exeAs extracts viable AS imperative from script
func exeAs(src *script.Script) (*script.AsCommand, error) {
	asCmds, ok := src.Preambles[script.CmdAs]
	if !ok {
		return nil, fmt.Errorf("Script missing valid %s", script.CmdAs)
	}
	asCmd := asCmds[0].(*script.AsCommand)
	return asCmd, nil
}
