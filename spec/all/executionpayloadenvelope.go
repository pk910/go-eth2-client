// Copyright © 2026 Attestant Limited.
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
	"errors"
	"fmt"

	"github.com/ethpandaops/go-eth2-client/spec/electra"
	"github.com/ethpandaops/go-eth2-client/spec/gloas"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/pk910/dynamic-ssz/sszutils"
)

// ExecutionPayloadEnvelope is a fork-agnostic execution payload envelope
// (EIP-7732 and later). Currently only gloas defines this container; later
// forks that share the schema will reuse the gloas view.
//
// There is no VersionedExecutionPayloadEnvelope wrapper in the spec package
// yet, so ToVersioned/FromVersioned are not provided. Add them once the
// Versioned* counterpart lands.
type ExecutionPayloadEnvelope struct {
	Version               version.DataVersion
	Payload               *ExecutionPayload
	ExecutionRequests     *electra.ExecutionRequests
	BuilderIndex          gloas.BuilderIndex
	BeaconBlockRoot       phase0.Root
	ParentBeaconBlockRoot phase0.Root
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (e *ExecutionPayloadEnvelope) viewType() (any, error) {
	switch e.Version {
	case version.DataVersionGloas:
		return (*gloas.ExecutionPayloadEnvelope)(nil), nil
	default:
		return nil, fmt.Errorf("ExecutionPayloadEnvelope: unsupported version %d", e.Version)
	}
}

// MarshalSSZDyn marshals the envelope using the view that matches Version.
func (e *ExecutionPayloadEnvelope) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := e.viewType()
	if err != nil {
		return nil, err
	}

	m, ok := any(e).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("ExecutionPayloadEnvelope: generated SSZ code missing")
	}

	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("ExecutionPayloadEnvelope: no view marshaler for version %d", e.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the envelope for the active Version.
func (e *ExecutionPayloadEnvelope) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
	view, err := e.viewType()
	if err != nil {
		return 0
	}

	s, ok := any(e).(sszutils.DynamicViewSizer)
	if !ok {
		return 0
	}

	fn := s.SizeSSZDynView(view)
	if fn == nil {
		return 0
	}

	return fn(ds)
}

// UnmarshalSSZDyn decodes the envelope into the view that matches Version.
func (e *ExecutionPayloadEnvelope) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := e.viewType()
	if err != nil {
		return err
	}

	u, ok := any(e).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("ExecutionPayloadEnvelope: generated SSZ code missing")
	}

	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("ExecutionPayloadEnvelope: no view unmarshaler for version %d", e.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	e.populateVersion(e.Version)

	return nil
}

// populateVersion sets Version and propagates it to nested versionable
// children allocated by the SSZ unmarshal.
func (e *ExecutionPayloadEnvelope) populateVersion(v version.DataVersion) {
	e.Version = v

	if e.Payload != nil {
		e.Payload.populateVersion(v)
	}
}

// ToView returns a fresh fork-specific ExecutionPayloadEnvelope populated
// with e's fields, recursing into Payload via copyByName.
func (e *ExecutionPayloadEnvelope) ToView() (any, error) {
	return toViewByCopy(e)
}

// FromView populates e from a fork-specific ExecutionPayloadEnvelope.
func (e *ExecutionPayloadEnvelope) FromView(view any) error {
	v, err := executionPayloadEnvelopeVersion(view)
	if err != nil {
		return err
	}

	if e.Version == version.DataVersionUnknown {
		e.Version = v
	}

	if err := copyByName(view, e); err != nil {
		return err
	}

	e.populateVersion(e.Version)

	return nil
}

// executionPayloadEnvelopeVersion maps an ExecutionPayloadEnvelope view type
// to its DataVersion.
func executionPayloadEnvelopeVersion(view any) (version.DataVersion, error) {
	switch view.(type) {
	case *gloas.ExecutionPayloadEnvelope:
		return version.DataVersionGloas, nil
	default:
		return version.DataVersionUnknown, fmt.Errorf("ExecutionPayloadEnvelope: unsupported view type %T", view)
	}
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (e *ExecutionPayloadEnvelope) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := e.viewType()
	if err != nil {
		return err
	}

	h, ok := any(e).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("ExecutionPayloadEnvelope: generated SSZ code missing")
	}

	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("ExecutionPayloadEnvelope: no view hasher for version %d", e.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (e *ExecutionPayloadEnvelope) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return e.MarshalSSZDyn(ds, make([]byte, 0, e.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (e *ExecutionPayloadEnvelope) MarshalSSZTo(dst []byte) ([]byte, error) {
	return e.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (e *ExecutionPayloadEnvelope) UnmarshalSSZ(buf []byte) error {
	return e.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (e *ExecutionPayloadEnvelope) SizeSSZ() int {
	return e.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (e *ExecutionPayloadEnvelope) HashTreeRoot() ([32]byte, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(e)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (e *ExecutionPayloadEnvelope) HashTreeRootWith(hh sszutils.HashWalker) error {
	return e.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}

// MarshalJSON delegates to the per-fork ExecutionPayloadEnvelope that matches Version.
func (e *ExecutionPayloadEnvelope) MarshalJSON() ([]byte, error) {
	return marshalAsView(e)
}

// UnmarshalJSON delegates to the per-fork ExecutionPayloadEnvelope that matches Version.
// Caller must set Version before calling.
func (e *ExecutionPayloadEnvelope) UnmarshalJSON(data []byte) error {
	if err := unmarshalAsView(e, data); err != nil {
		return err
	}

	e.populateVersion(e.Version)

	return nil
}

// MarshalYAML delegates to the per-fork ExecutionPayloadEnvelope that matches Version.
func (e *ExecutionPayloadEnvelope) MarshalYAML() ([]byte, error) {
	return marshalAsViewYAML(e)
}

// UnmarshalYAML delegates to the per-fork ExecutionPayloadEnvelope that matches Version.
// Caller must set Version before calling.
func (e *ExecutionPayloadEnvelope) UnmarshalYAML(data []byte) error {
	if err := unmarshalAsViewYAML(e, data); err != nil {
		return err
	}

	e.populateVersion(e.Version)

	return nil
}
