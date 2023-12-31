// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocompile

import (
	"fmt"

	"liummeng07/gocompile/buildcfg"
	"liummeng07/gocompile/internal/amd64"
	"liummeng07/gocompile/internal/arm"
	"liummeng07/gocompile/internal/arm64"
	"liummeng07/gocompile/internal/base"
	"liummeng07/gocompile/internal/gc"
	"liummeng07/gocompile/internal/mips"
	"liummeng07/gocompile/internal/mips64"
	"liummeng07/gocompile/internal/ppc64"
	"liummeng07/gocompile/internal/riscv64"
	"liummeng07/gocompile/internal/s390x"
	"liummeng07/gocompile/internal/ssagen"
	"liummeng07/gocompile/internal/wasm"
	"liummeng07/gocompile/internal/x86"
	"log"
	"os"
)

var archInits = map[string]func(*ssagen.ArchInfo){
	"386":      x86.Init,
	"amd64":    amd64.Init,
	"arm":      arm.Init,
	"arm64":    arm64.Init,
	"mips":     mips.Init,
	"mipsle":   mips.Init,
	"mips64":   mips64.Init,
	"mips64le": mips64.Init,
	"ppc64":    ppc64.Init,
	"ppc64le":  ppc64.Init,
	"riscv64":  riscv64.Init,
	"s390x":    s390x.Init,
	"wasm":     wasm.Init,
}

func main() {
	// disable timestamps for reproducible output
	log.SetFlags(0)
	log.SetPrefix("compile: ")

	buildcfg.Check()
	archInit, ok := archInits[buildcfg.GOARCH]
	if !ok {
		fmt.Fprintf(os.Stderr, "compile: unknown architecture %q\n", buildcfg.GOARCH)
		os.Exit(2)
	}

	gc.Main(archInit)
	base.Exit(0)
}
