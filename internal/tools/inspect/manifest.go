// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package inspect

import (
	"debug/dwarf"
	"errors"
	"fmt"
	"io"

	"github.com/hashicorp/go-version"
)

// Manifest contains all information that needs to be inspected for an
// application.
type Manifest struct {
	// Application is the application to extract binary data from.
	Application Application

	// Packages are the package declarations to inspect.
	Packages []Package
	// StructFields are struct fields the application should contain that need
	// offsets to be found.
	StructFields []StructField
}

func (m Manifest) validate() error {
	if m.Application.GoVerions == nil && m.Application.Versions == nil {
		return errors.New("missing version: a Go or application version is required")
	}
	return nil
}

// Application is the information about a template application that needs to be
// inspected for binary data.
type Application struct {
	// Renderer renders the application.
	Renderer Renderer
	// Versions are the application versions to be inspected. They will be
	// passed to the Renderer as the ".Version" field.
	//
	// If this is nil, the GoVerions will also be used as the application
	// versions that are passed to the template.
	Versions []*version.Version
	// GoVerions are the versions of Go to build the application with.
	//
	// If this is nil, the latest version of Go will be used.
	GoVerions []*version.Version
}

// Package contains all the declarations from a package to inspect.
type Package struct {
	// ImportPath is the unique import path of the package.
	ImportPath string

	// Structs are the structs within the package to inspect.
	Structs []Struct

	// TODO: If functions are every instrumented (i.e. not methods) they can be
	// added here.
}

// funcs returns the fully-qualified names of all the instrumented functions
// (including methods) within a package. The package names are constructed to
// match the DWARF symbol names of a Go package.
func (p Package) funcs() []string {
	var out []string
	for _, s := range p.Structs {
		out = append(out, s.funcs(p.ImportPath)...)
	}
	return out
}

func (p Package) structFields() []StructField {
	var out []StructField
	for _, s := range p.Structs {
		for _, f := range s.Fields {
			out = append(out, StructField{
				PkgPath: p.ImportPath,
				Struct:  s.Name,
				Field:   f.Name,
			})
		}
	}
	return out
}

// Struct is a Go struct type within a package that to inspect.
type Struct struct {
	// Name is the struct type name.
	Name string
	// Fields are the fields within the struct to inspect.
	Fields []Field
	// Methods are the struct methods to inspect.
	Methods []Method
}

func (s Struct) funcs(pkg string) []string {
	var out []string
	for _, m := range s.Methods {
		out = append(out, m.name(pkg, s.Name))
	}
	return out
}

// Method is a method of a struct within a package being inspected.
type Method struct {
	// Name is the method name.
	Name string
	// Indirect defines if the method is called with an indirect receiver.
	Indirect bool
}

// name returns the fully-qualified method name.
func (m Method) name(pkg, strct string) string {
	if m.Indirect {
		return fmt.Sprintf("%s.(*%s).%s", pkg, strct, m.Name)
	}
	return fmt.Sprintf("%s.%s.%s", pkg, strct, m.Name)
}

// Field is a field of a struct within a package being inspected.
type Field struct {
	// Name is the field name.
	Name string
}

// StructField defines a field of a struct from a package.
type StructField struct {
	// PkgPath is the unique package import path containing the struct.
	PkgPath string
	// Struct is the name of the struct containing the field.
	Struct string
	// Field is the name of the field.
	Field string
}

// structName returns the package path prefixed struct name of s.
func (s StructField) structName() string {
	return fmt.Sprintf("%s.%s", s.PkgPath, s.Struct)
}

// offset returns the field offset found in the DWARF data d and true. If the
// offset is not found in d, 0 and false are returned.
func (s StructField) offset(d *dwarf.Data) (uint64, bool) {
	r := d.Reader()
	if !gotoEntry(r, dwarf.TagStructType, s.structName()) {
		return 0, false
	}

	e, err := findEntry(r, dwarf.TagMember, s.Field)
	if err != nil {
		return 0, false
	}

	f, ok := entryField(e, dwarf.AttrDataMemberLoc)
	if !ok {
		return 0, false
	}

	return uint64(f.Val.(int64)), true
}

// gotoEntry reads from r until the entry with a tag equal to name is found.
// True is returned if the entry is found, otherwise false is returned.
func gotoEntry(r *dwarf.Reader, tag dwarf.Tag, name string) bool {
	_, err := findEntry(r, tag, name)
	return err == nil
}

var errNotFound = errors.New("not found")

// findEntry returns the DWARF entry with a tag equal to name read from r. An
// error is returned if the entry cannot be found.
func findEntry(r *dwarf.Reader, tag dwarf.Tag, name string) (*dwarf.Entry, error) {
	for {
		entry, err := r.Next()
		if err == io.EOF || entry == nil {
			break
		}

		if entry.Tag == tag {
			if f, ok := entryField(entry, dwarf.AttrName); ok {
				if name == f.Val.(string) {
					return entry, nil
				}
			}
		}
	}
	return nil, errNotFound
}

// entryField returns the DWARF field from DWARF entry e that has the passed
// DWARF attribute a.
func entryField(e *dwarf.Entry, a dwarf.Attr) (dwarf.Field, bool) {
	for _, f := range e.Field {
		if f.Attr == a {
			return f, true
		}
	}
	return dwarf.Field{}, false
}
