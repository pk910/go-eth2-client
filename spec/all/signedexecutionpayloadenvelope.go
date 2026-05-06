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

	"github.com/ethpandaops/go-eth2-client/spec/gloas"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/pk910/dynamic-ssz/sszutils"
)

// SignedExecutionPayloadEnvelope is a fork-agnostic signed execution payload
// envelope. Currently only gloas defines this container; later forks that
// share the schema will reuse the gloas view.
//
// There is no VersionedSignedExecutionPayloadEnvelope wrapper in the spec
// package yet, so ToVersioned/FromVersioned are not provided. Add them once
// the Versioned* counterpart lands.
type SignedExecutionPayloadEnvelope struct {
	Version   version.DataVersion
	Message   *ExecutionPayloadEnvelope
	Signature phase0.BLSSignature
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (s *SignedExecutionPayloadEnvelope) viewType() (any, error) {
	switch s.Version {
	case version.DataVersionGloas:
		return (*gloas.SignedExecutionPayloadEnvelope)(nil), nil
	default:
		return nil, fmt.Errorf("SignedExecutionPayloadEnvelope: unsupported version %d", s.Version)
	}
}

// MarshalSSZDyn marshals the signed envelope using the view that matches Version.
func (s *SignedExecutionPayloadEnvelope) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := s.viewType()
	if err != nil {
		return nil, err
	}

	m, ok := any(s).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("SignedExecutionPayloadEnvelope: generated SSZ code missing")
	}

	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("SignedExecutionPayloadEnvelope: no view marshaler for version %d", s.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the signed envelope for the active Version.
func (s *SignedExecutionPayloadEnvelope) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
	view, err := s.viewType()
	if err != nil {
		return 0
	}

	sz, ok := any(s).(sszutils.DynamicViewSizer)
	if !ok {
		return 0
	}

	fn := sz.SizeSSZDynView(view)
	if fn == nil {
		return 0
	}

	return fn(ds)
}

// UnmarshalSSZDyn decodes the signed envelope into the view that matches Version.
func (s *SignedExecutionPayloadEnvelope) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := s.viewType()
	if err != nil {
		return err
	}

	u, ok := any(s).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("SignedExecutionPayloadEnvelope: generated SSZ code missing")
	}

	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("SignedExecutionPayloadEnvelope: no view unmarshaler for version %d", s.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	s.populateVersion(s.Version)

	return nil
}

// populateVersion sets Version and propagates it to the inner message.
func (s *SignedExecutionPayloadEnvelope) populateVersion(v version.DataVersion) {
	s.Version = v

	if s.Message != nil {
		s.Message.populateVersion(v)
	}
}

// ToView returns a fresh fork-specific SignedExecutionPayloadEnvelope
// populated with s's fields, recursing into Message via copyByName.
func (s *SignedExecutionPayloadEnvelope) ToView() (any, error) {
	return toViewByCopy(s)
}

// FromView populates s from a fork-specific SignedExecutionPayloadEnvelope.
func (s *SignedExecutionPayloadEnvelope) FromView(view any) error {
	v, err := signedExecutionPayloadEnvelopeVersion(view)
	if err != nil {
		return err
	}

	if s.Version == version.DataVersionUnknown {
		s.Version = v
	}

	if err := copyByName(view, s); err != nil {
		return err
	}

	s.populateVersion(s.Version)

	return nil
}

// signedExecutionPayloadEnvelopeVersion maps a SignedExecutionPayloadEnvelope
// view type to its DataVersion.
func signedExecutionPayloadEnvelopeVersion(view any) (version.DataVersion, error) {
	switch view.(type) {
	case *gloas.SignedExecutionPayloadEnvelope:
		return version.DataVersionGloas, nil
	default:
		return version.DataVersionUnknown, fmt.Errorf("SignedExecutionPayloadEnvelope: unsupported view type %T", view)
	}
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (s *SignedExecutionPayloadEnvelope) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := s.viewType()
	if err != nil {
		return err
	}

	h, ok := any(s).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("SignedExecutionPayloadEnvelope: generated SSZ code missing")
	}

	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("SignedExecutionPayloadEnvelope: no view hasher for version %d", s.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (s *SignedExecutionPayloadEnvelope) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return s.MarshalSSZDyn(ds, make([]byte, 0, s.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (s *SignedExecutionPayloadEnvelope) MarshalSSZTo(dst []byte) ([]byte, error) {
	return s.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (s *SignedExecutionPayloadEnvelope) UnmarshalSSZ(buf []byte) error {
	return s.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (s *SignedExecutionPayloadEnvelope) SizeSSZ() int {
	return s.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (s *SignedExecutionPayloadEnvelope) HashTreeRoot() ([32]byte, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(s)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (s *SignedExecutionPayloadEnvelope) HashTreeRootWith(hh sszutils.HashWalker) error {
	return s.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}

// MarshalJSON delegates to the per-fork SignedExecutionPayloadEnvelope that matches Version.
func (s *SignedExecutionPayloadEnvelope) MarshalJSON() ([]byte, error) {
	return marshalAsView(s)
}

// UnmarshalJSON delegates to the per-fork SignedExecutionPayloadEnvelope that matches Version.
// Caller must set Version before calling.
func (s *SignedExecutionPayloadEnvelope) UnmarshalJSON(data []byte) error {
	if err := unmarshalAsView(s, data); err != nil {
		return err
	}

	s.populateVersion(s.Version)

	return nil
}

// MarshalYAML delegates to the per-fork SignedExecutionPayloadEnvelope that matches Version.
func (s *SignedExecutionPayloadEnvelope) MarshalYAML() ([]byte, error) {
	return marshalAsViewYAML(s)
}

// UnmarshalYAML delegates to the per-fork SignedExecutionPayloadEnvelope that matches Version.
// Caller must set Version before calling.
func (s *SignedExecutionPayloadEnvelope) UnmarshalYAML(data []byte) error {
	if err := unmarshalAsViewYAML(s, data); err != nil {
		return err
	}

	s.populateVersion(s.Version)

	return nil
}
