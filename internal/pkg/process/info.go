// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package process

import (
	"debug/elf"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/Masterminds/semver/v3"

	"go.opentelemetry.io/auto/internal/pkg/process/binary"
)

// Info are the details about a target process.
type Info struct {
	PID       int
	Functions *Functions
	GoVersion *semver.Version
	Modules   map[string]*semver.Version

	allocOnce onceResult[*Allocation]
}

// NewInfo loads the process Info for pid. Only functions whose names return
// true from fnFilter are loaded in the returned Info.
func NewInfo(pid int, fnFilter func(string) bool) (*Info, error) {
	id := ID(pid)

	path, err := id.ExecPath()
	if err != nil {
		return nil, err
	}
	elfF, err := elf.Open(path)
	if err != nil {
		return nil, err
	}
	defer elfF.Close()

	bi, err := id.BuildInfo()
	if err != nil {
		return nil, err
	}

	goVersion, err := semver.NewVersion(bi.GoVersion)
	if err != nil {
		return nil, err
	}

	result := &Info{
		PID:       pid,
		GoVersion: goVersion,
	}

	result.Modules, err = findModules(goVersion, bi.Deps)
	var e error
	result.Functions, e = loadFunctions(elfF, fnFilter)
	err = errors.Join(err, e)

	return result, err
}

func findModules(goVer *semver.Version, deps []*debug.Module) (map[string]*semver.Version, error) {
	var err error
	out := make(map[string]*semver.Version, len(deps)+1)
	for _, dep := range deps {
		depVersion, e := semver.NewVersion(dep.Version)
		if e != nil {
			err = errors.Join(
				err,
				fmt.Errorf("invalid dependency version %s (%s): %w", dep.Path, dep.Version, e),
			)
			continue
		}
		out[dep.Path] = depVersion
	}
	out["std"] = goVer
	return out, err
}

// Alloc allocates memory for the process defined by i. This method only makes
// a single allocation on the first successful call. All subsequent calls will
// receive the Allocation information for that call.
func (i *Info) Alloc() (*Allocation, error) {
	return i.allocOnce.Do(func() (*Allocation, error) {
		// TODO: fix logger arg
		return Allocate(nil, i.PID)
	})
}

// Functions are Go functions found in a process.
type Functions struct {
	fn map[string]*binary.Func
}

func NewFunctions() *Functions {
	return &Functions{fn: make(map[string]*binary.Func)}
}

// LoadFunctions returns Functions containing all the functions found in the
// process identified by pid and filtered by fltr. If fltr returns true for a
// passed function name the function will be included in the returned Functions,
// otherwise it will be dropped.
func LoadFunctions(pid int, fltr func(string) bool) (*Functions, error) {
	id := ID(pid)
	path, err := id.ExecPath()
	if err != nil {
		return nil, err
	}
	elfF, err := elf.Open(path)
	if err != nil {
		return nil, err
	}
	defer elfF.Close()
	return loadFunctions(elfF, fltr)
}

func loadFunctions(elfF *elf.File, fltr func(string) bool) (*Functions, error) {
	if fltr == nil {
		fltr = func(string) bool { return true }
	}

	found, err := binary.FindFunctionsUnStripped(elfF, fltr)
	if err != nil {
		if !errors.Is(err, elf.ErrNoSymbols) {
			return nil, err
		}
		found, err = binary.FindFunctionsStripped(elfF, fltr)
		if err != nil {
			return nil, err
		}
	}

	if len(found) == 0 {
		return nil, errors.New("no functions found")
	}

	return &Functions{fn: found}, nil
}

// Len returns the number of functions contained in f.
func (f *Functions) Len() int {
	return len(f.fn)
}

// Put adds fn to Functions. Any existing function with the same name will be
// overwritten.
func (f *Functions) Put(fn *binary.Func) {
	f.fn[fn.Name] = fn
}

// Get returns the function with the provided name and true if it exists.
// Otherwise, nil and false are returned.
func (f *Functions) Get(name string) (*binary.Func, bool) {
	got, ok := f.fn[name]
	return got, ok
}

func (f *Functions) get(name string) (*binary.Func, error) {
	got, ok := f.Get(name)
	if !ok {
		return nil, fmt.Errorf("unknown function: %s", name)
	}
	return got, nil
}

// Offset returns the offset for the function with name.
func (f *Functions) Offset(name string) (uint64, error) {
	got, err := f.get(name)
	if err != nil {
		return 0, err
	}
	return got.Offset, nil
}

// ReturnOffsets returns the return offset values for the function with name.
func (f *Functions) ReturnOffsets(name string) ([]uint64, error) {
	got, err := f.get(name)
	if err != nil {
		return nil, err
	}
	return got.ReturnOffsets, nil
}

// onceResult is an object that will perform exactly one action if that action
// does not error. For errors, no state is stored and subsequent attempts will
// be tried.
type onceResult[T any] struct {
	done  atomic.Bool
	mutex sync.Mutex
	val   T
}

// Do runs f only once, and only stores the result if f returns a nil error.
// Subsequent calls to Do will return the stored value or they will re-attempt
// to run f and store the result if an error had been returned.
func (o *onceResult[T]) Do(f func() (T, error)) (T, error) {
	if o.done.Load() {
		o.mutex.Lock()
		defer o.mutex.Unlock()
		return o.val, nil
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.done.Load() {
		return o.val, nil
	}

	var err error
	o.val, err = f()
	if err == nil {
		o.done.Store(true)
	}
	return o.val, err
}
