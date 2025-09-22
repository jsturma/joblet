//go:build tools
// +build tools

// This file ensures that build-time dependencies are tracked in go.mod
// even though they're not imported in regular Go code.

package main

import (
	_ "github.com/ehsaniara/joblet-proto"
)
