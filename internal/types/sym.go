// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"liummeng07/gocompile/internal/base"
	"liummeng07/gocompile/obj"
	"liummeng07/gocompile/src"
	"unicode"
	"unicode/utf8"
)

// Sym represents an object name in a segmented (pkg, name) namespace.
// Most commonly, this is a Go identifier naming an object declared within a package,
// but Syms are also used to name internal synthesized objects.
//
// As an exception, field and method names that are exported use the Sym
// associated with localpkg instead of the package that declared them. This
// allows using Sym pointer equality to test for Go identifier uniqueness when
// handling selector expressions.
//
// Ideally, Sym should be used for representing Go language constructs,
// while liummeng07/gocompile/obj.LSym is used for representing emitted artifacts.
//
// NOTE: In practice, things can be messier than the description above
// for various reasons (historical, convenience).
type Sym struct {
	Linkname string // link name

	Pkg  *Pkg
	Name string // object name

	// Def, Block, and Lastlineno are saved and restored by Pushdcl/Popdcl.

	// The unique ONAME, OTYPE, OPACK, or OLITERAL node that this symbol is
	// bound to within the current scope. (Most parts of the compiler should
	// prefer passing the Node directly, rather than relying on this field.)
	Def        Object
	Block      int32    // blocknumber to catch redeclaration
	Lastlineno src.XPos // last declaration for diagnostic

	flags bitset8
}

const (
	symOnExportList = 1 << iota // added to exportlist (no need to add again)
	symUniq
	symSiggen // type symbol has been generated
	symAsm    // on asmlist, for writing to -asmhdr
	symFunc   // function symbol
)

func (sym *Sym) OnExportList() bool { return sym.flags&symOnExportList != 0 }
func (sym *Sym) Uniq() bool         { return sym.flags&symUniq != 0 }
func (sym *Sym) Siggen() bool       { return sym.flags&symSiggen != 0 }
func (sym *Sym) Asm() bool          { return sym.flags&symAsm != 0 }
func (sym *Sym) Func() bool         { return sym.flags&symFunc != 0 }

func (sym *Sym) SetOnExportList(b bool) { sym.flags.set(symOnExportList, b) }
func (sym *Sym) SetUniq(b bool)         { sym.flags.set(symUniq, b) }
func (sym *Sym) SetSiggen(b bool)       { sym.flags.set(symSiggen, b) }
func (sym *Sym) SetAsm(b bool)          { sym.flags.set(symAsm, b) }
func (sym *Sym) SetFunc(b bool)         { sym.flags.set(symFunc, b) }

func (sym *Sym) IsBlank() bool {
	return sym != nil && sym.Name == "_"
}

// Deprecated: This method should not be used directly. Instead, use a
// higher-level abstraction that directly returns the linker symbol
// for a named object. For example, reflectdata.TypeLinksym(t) instead
// of reflectdata.TypeSym(t).Linksym().
func (sym *Sym) Linksym() *obj.LSym {
	abi := obj.ABI0
	if sym.Func() {
		abi = obj.ABIInternal
	}
	return sym.LinksymABI(abi)
}

// Deprecated: This method should not be used directly. Instead, use a
// higher-level abstraction that directly returns the linker symbol
// for a named object. For example, (*ir.Name).LinksymABI(abi) instead
// of (*ir.Name).Sym().LinksymABI(abi).
func (sym *Sym) LinksymABI(abi obj.ABI) *obj.LSym {
	if sym == nil {
		base.Fatalf("nil symbol")
	}
	if sym.Linkname != "" {
		return base.Linkname(sym.Linkname, abi)
	}
	return base.PkgLinksym(sym.Pkg.Prefix, sym.Name, abi)
}

// Less reports whether symbol a is ordered before symbol b.
//
// Symbols are ordered exported before non-exported, then by name, and
// finally (for non-exported symbols) by package height and path.
//
// Ordering by package height is necessary to establish a consistent
// ordering for non-exported names with the same spelling but from
// different packages. We don't necessarily know the path for the
// package being compiled, but by definition it will have a height
// greater than any other packages seen within the compilation unit.
// For more background, see issue #24693.
func (a *Sym) Less(b *Sym) bool {
	if a == b {
		return false
	}

	// Nil before non-nil.
	if a == nil {
		return true
	}
	if b == nil {
		return false
	}

	// Exported symbols before non-exported.
	ea := IsExported(a.Name)
	eb := IsExported(b.Name)
	if ea != eb {
		return ea
	}

	// Order by name and then (for non-exported names) by package
	// height and path.
	if a.Name != b.Name {
		return a.Name < b.Name
	}
	if !ea {
		if a.Pkg.Height != b.Pkg.Height {
			return a.Pkg.Height < b.Pkg.Height
		}
		return a.Pkg.Path < b.Pkg.Path
	}
	return false
}

// IsExported reports whether name is an exported Go symbol (that is,
// whether it begins with an upper-case letter).
func IsExported(name string) bool {
	if r := name[0]; r < utf8.RuneSelf {
		return 'A' <= r && r <= 'Z'
	}
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}
