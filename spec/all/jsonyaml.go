// Copyright © 2023 Attestant Limited.
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

package all

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/goccy/go-yaml"
)

// viewProvider is implemented by fork-agnostic union types that can map their
// active Version to a fork-specific schema type via viewType().
type viewProvider interface {
	viewType() (any, error)
}

// marshalAsView delegates JSON marshaling to the per-fork type that matches
// src.Version. It allocates a fresh fork-specific instance, copies matching
// fields from src into it via copyByName, and lets json.Marshal route through
// the per-fork type's MarshalJSON.
func marshalAsView(src viewProvider) ([]byte, error) {
	inst, err := newViewInstance(src)
	if err != nil {
		return nil, err
	}

	if err := copyByName(src, inst); err != nil {
		return nil, err
	}

	return json.Marshal(inst)
}

// unmarshalAsView delegates JSON unmarshaling to the per-fork type matching
// dst.Version, then copies the populated view back into dst. The caller is
// responsible for invoking populateVersion afterwards to seed the version on
// any nested versionable children that copyByName allocated.
func unmarshalAsView(dst viewProvider, data []byte) error {
	inst, err := newViewInstance(dst)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, inst); err != nil {
		return err
	}

	return copyByName(inst, dst)
}

// marshalAsViewYAML mirrors marshalAsView for YAML.
func marshalAsViewYAML(src viewProvider) ([]byte, error) {
	inst, err := newViewInstance(src)
	if err != nil {
		return nil, err
	}

	if err := copyByName(src, inst); err != nil {
		return nil, err
	}

	return yaml.Marshal(inst)
}

// unmarshalAsViewYAML mirrors unmarshalAsView for YAML.
func unmarshalAsViewYAML(dst viewProvider, data []byte) error {
	inst, err := newViewInstance(dst)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, inst); err != nil {
		return err
	}

	return copyByName(inst, dst)
}

// newViewInstance allocates a fresh instance of the fork-specific schema type
// indicated by p.viewType().
func newViewInstance(p viewProvider) (any, error) {
	view, err := p.viewType()
	if err != nil {
		return nil, err
	}

	return reflect.New(reflect.TypeOf(view).Elem()).Interface(), nil
}

// copyByName copies fields from src to dst by matching field names. Both
// arguments must be pointers to structs. For directly-assignable types
// (same Go type on both sides) the field is assigned. For pointer or slice
// fields where src/dst element types differ — *all.X / *phase0.X, slices
// of those, etc. — copyByName recursively allocates and copies. Fields
// present on src but missing on dst (e.g. Version when dst is per-fork,
// or BaseFeePerGasLE when dst is a deneb-style view) are silently skipped.
func copyByName(src, dst any) error {
	sv := reflect.ValueOf(src)
	dv := reflect.ValueOf(dst)

	if sv.Kind() == reflect.Ptr {
		if sv.IsNil() {
			return nil
		}

		sv = sv.Elem()
	}

	if dv.Kind() == reflect.Ptr {
		if dv.IsNil() {
			return errors.New("copyByName: destination is a nil pointer")
		}

		dv = dv.Elem()
	}

	if sv.Kind() != reflect.Struct || dv.Kind() != reflect.Struct {
		return fmt.Errorf("copyByName: src kind=%s, dst kind=%s; both must be structs", sv.Kind(), dv.Kind())
	}

	for i := range dv.NumField() {
		df := dv.Type().Field(i)
		if !df.IsExported() {
			continue
		}

		sf := sv.FieldByName(df.Name)
		if !sf.IsValid() {
			continue
		}

		if err := copyValue(sf, dv.Field(i)); err != nil {
			return fmt.Errorf("field %s: %w", df.Name, err)
		}
	}

	return nil
}

// copyValue copies a single field value, recursing into nested struct
// pointers and slices when types differ. Direct assignment is preferred
// when src.Type() is assignable to dst.Type().
func copyValue(src, dst reflect.Value) error {
	if src.Type().AssignableTo(dst.Type()) {
		dst.Set(src)

		return nil
	}

	switch src.Kind() {
	case reflect.Ptr:
		if dst.Kind() != reflect.Ptr {
			return fmt.Errorf("incompatible kinds %s -> %s", src.Kind(), dst.Kind())
		}

		if src.IsNil() {
			dst.Set(reflect.Zero(dst.Type()))

			return nil
		}

		newDst := reflect.New(dst.Type().Elem())
		if err := copyByName(src.Interface(), newDst.Interface()); err != nil {
			return err
		}

		dst.Set(newDst)

		return nil
	case reflect.Slice:
		if dst.Kind() != reflect.Slice {
			return fmt.Errorf("incompatible kinds %s -> %s", src.Kind(), dst.Kind())
		}

		if src.IsNil() {
			dst.Set(reflect.Zero(dst.Type()))

			return nil
		}

		n := src.Len()
		newDst := reflect.MakeSlice(dst.Type(), n, n)

		for i := range n {
			if err := copyValue(src.Index(i), newDst.Index(i)); err != nil {
				return err
			}
		}

		dst.Set(newDst)

		return nil
	default:
		return fmt.Errorf("cannot copy %s to %s", src.Type(), dst.Type())
	}
}
