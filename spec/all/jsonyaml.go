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

	"github.com/ethpandaops/go-eth2-client/spec/version"
	"github.com/goccy/go-yaml"
)

// viewProvider is implemented by fork-agnostic union types that can map their
// active Version to a fork-specific schema type via viewType().
type viewProvider interface {
	viewType() (any, error)
}

// viewer is implemented by every fork-agnostic union type's ToView method.
type viewer interface {
	ToView() (any, error)
}

// fromViewer is implemented by every fork-agnostic union type's FromView method.
type fromViewer interface {
	FromView(view any) error
}

// versionSetter is implemented by every fork-agnostic union type's
// populateVersion method. fromVersioned uses it to pin the Version (taken
// from the Versioned* struct's authoritative Version field) before FromView
// runs — otherwise FromView's per-view-type inference loses precision when
// multiple versions share a single view (e.g. phase0..deneb all use
// *phase0.Attestation).
type versionSetter interface {
	populateVersion(v version.DataVersion)
}

// versionFieldName maps a DataVersion to the per-fork field name on a
// Versioned* struct (e.g. *spec.VersionedBeaconState).
func versionFieldName(v version.DataVersion) (string, error) {
	switch v {
	case version.DataVersionPhase0:
		return "Phase0", nil
	case version.DataVersionAltair:
		return "Altair", nil
	case version.DataVersionBellatrix:
		return "Bellatrix", nil
	case version.DataVersionCapella:
		return "Capella", nil
	case version.DataVersionDeneb:
		return "Deneb", nil
	case version.DataVersionElectra:
		return "Electra", nil
	case version.DataVersionFulu:
		return "Fulu", nil
	case version.DataVersionGloas:
		return "Gloas", nil
	case version.DataVersionHeze:
		return "Heze", nil
	default:
		return "", fmt.Errorf("unsupported version %d", v)
	}
}

// toVersioned populates a Versioned* struct from a fork-agnostic source.
// It calls src.ToView to produce the fork-specific view and stores it in the
// dst field whose name matches srcVersion (Phase0, Altair, …, Heze). The
// Version field on dst is also set.
//
// dst must be a non-nil pointer to a struct shaped like the spec package's
// Versioned* types: a Version field plus per-fork pointer fields.
func toVersioned(srcVersion version.DataVersion, src viewer, dst any) error {
	view, err := src.ToView()
	if err != nil {
		return err
	}

	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return errors.New("toVersioned: dst must be a non-nil pointer")
	}

	dv = dv.Elem()

	versionField := dv.FieldByName("Version")
	if !versionField.IsValid() {
		return fmt.Errorf("toVersioned: %T has no Version field", dst)
	}

	versionField.Set(reflect.ValueOf(srcVersion))

	fieldName, err := versionFieldName(srcVersion)
	if err != nil {
		return fmt.Errorf("toVersioned: %w", err)
	}

	f := dv.FieldByName(fieldName)
	if !f.IsValid() {
		return fmt.Errorf("toVersioned: %T has no %s field", dst, fieldName)
	}

	rv := reflect.ValueOf(view)
	if !rv.Type().AssignableTo(f.Type()) {
		return fmt.Errorf("toVersioned: view type %T not assignable to %s field of type %s",
			view, fieldName, f.Type())
	}

	f.Set(rv)

	return nil
}

// fromVersioned populates a fork-agnostic destination from a Versioned*
// struct by extracting the field matching src.Version and feeding it to
// dst.FromView.
//
// src must be a non-nil pointer to a struct shaped like the spec package's
// Versioned* types.
func fromVersioned(dst fromViewer, src any) error {
	sv := reflect.ValueOf(src)
	if sv.Kind() != reflect.Ptr || sv.IsNil() {
		return errors.New("fromVersioned: src must be a non-nil pointer")
	}

	sv = sv.Elem()

	versionField := sv.FieldByName("Version")
	if !versionField.IsValid() {
		return fmt.Errorf("fromVersioned: %T has no Version field", src)
	}

	v, ok := versionField.Interface().(version.DataVersion)
	if !ok {
		return fmt.Errorf("fromVersioned: Version field on %T is %T not version.DataVersion",
			src, versionField.Interface())
	}

	fieldName, err := versionFieldName(v)
	if err != nil {
		return fmt.Errorf("fromVersioned: %w", err)
	}

	f := sv.FieldByName(fieldName)
	if !f.IsValid() {
		return fmt.Errorf("fromVersioned: %T has no %s field", src, fieldName)
	}

	if f.Kind() == reflect.Ptr && f.IsNil() {
		return fmt.Errorf("fromVersioned: %T.%s is nil for Version=%s", src, fieldName, v)
	}

	// Pin the authoritative Version on dst before FromView, so FromView's
	// type-switch inference doesn't downgrade it (e.g. *phase0.Attestation
	// would otherwise always become Phase0 even when the source Versioned
	// structure says Deneb).
	if vs, ok := dst.(versionSetter); ok {
		vs.populateVersion(v)
	}

	return dst.FromView(f.Interface())
}

// toViewByCopy is the generic implementation of ToView shared by every
// fork-agnostic union type. It allocates a fresh fork-specific instance for
// the active Version and field-copies src into it via copyByName. copyByName
// recursively handles nested struct pointers and slices, so all fork-specific
// children — including nested *all.X via their own field structure — are
// populated transparently.
func toViewByCopy(src viewProvider) (any, error) {
	inst, err := newViewInstance(src)
	if err != nil {
		return nil, err
	}

	if err := copyByName(src, inst); err != nil {
		return nil, err
	}

	return inst, nil
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
