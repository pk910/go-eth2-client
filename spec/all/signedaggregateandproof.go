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
	"errors"
	"fmt"

	"github.com/ethpandaops/go-eth2-client/spec/electra"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/pk910/dynamic-ssz/sszutils"
)

// SignedAggregateAndProof is a fork-agnostic signed aggregate and proof.
type SignedAggregateAndProof struct {
	Version   version.DataVersion
	Message   *AggregateAndProof
	Signature phase0.BLSSignature
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (s *SignedAggregateAndProof) viewType() (any, error) {
	switch s.Version {
	case version.DataVersionPhase0,
		version.DataVersionAltair,
		version.DataVersionBellatrix,
		version.DataVersionCapella,
		version.DataVersionDeneb:
		return (*phase0.SignedAggregateAndProof)(nil), nil
	case version.DataVersionElectra,
		version.DataVersionFulu,
		version.DataVersionGloas,
		version.DataVersionHeze:
		return (*electra.SignedAggregateAndProof)(nil), nil
	default:
		return nil, fmt.Errorf("SignedAggregateAndProof: unsupported version %d", s.Version)
	}
}

// MarshalSSZDyn marshals the signed proof using the view that matches Version.
func (s *SignedAggregateAndProof) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := s.viewType()
	if err != nil {
		return nil, err
	}

	m, ok := any(s).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("SignedAggregateAndProof: generated SSZ code missing")
	}

	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("SignedAggregateAndProof: no view marshaler for version %d", s.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the signed proof for the active Version.
func (s *SignedAggregateAndProof) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
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

// UnmarshalSSZDyn decodes the signed proof into the view that matches Version.
func (s *SignedAggregateAndProof) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := s.viewType()
	if err != nil {
		return err
	}

	u, ok := any(s).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("SignedAggregateAndProof: generated SSZ code missing")
	}

	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("SignedAggregateAndProof: no view unmarshaler for version %d", s.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	s.populateVersion(s.Version)

	return nil
}

// populateVersion sets Version and propagates it to the inner message.
func (s *SignedAggregateAndProof) populateVersion(v version.DataVersion) {
	s.Version = v

	if s.Message != nil {
		s.Message.populateVersion(v)
	}
}

// ToView returns a fresh fork-specific SignedAggregateAndProof populated with
// s's fields, recursing into Message via its ToView.
func (s *SignedAggregateAndProof) ToView() (any, error) {
	var msg any

	var err error

	if s.Message != nil {
		msg, err = s.Message.ToView()
		if err != nil {
			return nil, err
		}
	}

	switch s.Version {
	case version.DataVersionPhase0,
		version.DataVersionAltair,
		version.DataVersionBellatrix,
		version.DataVersionCapella,
		version.DataVersionDeneb:
		pm, err := assertView[*phase0.AggregateAndProof](msg, "SignedAggregateAndProof.Message")
		if err != nil {
			return nil, err
		}

		return &phase0.SignedAggregateAndProof{Message: pm, Signature: s.Signature}, nil
	case version.DataVersionElectra,
		version.DataVersionFulu,
		version.DataVersionGloas,
		version.DataVersionHeze:
		em, err := assertView[*electra.AggregateAndProof](msg, "SignedAggregateAndProof.Message")
		if err != nil {
			return nil, err
		}

		return &electra.SignedAggregateAndProof{Message: em, Signature: s.Signature}, nil
	default:
		return nil, fmt.Errorf("SignedAggregateAndProof: unsupported version %d", s.Version)
	}
}

// FromView populates s from a fork-specific SignedAggregateAndProof.
func (s *SignedAggregateAndProof) FromView(view any) error {
	var msgView any

	switch v := view.(type) {
	case *phase0.SignedAggregateAndProof:
		if s.Version == version.DataVersionUnknown {
			s.Version = version.DataVersionPhase0
		}

		s.Signature = v.Signature
		msgView = v.Message
	case *electra.SignedAggregateAndProof:
		if s.Version == version.DataVersionUnknown {
			s.Version = version.DataVersionElectra
		}

		s.Signature = v.Signature

		if v.Message != nil {
			msgView = v.Message
		}
	default:
		return fmt.Errorf("SignedAggregateAndProof: unsupported view type %T", view)
	}

	if msgView == nil {
		s.Message = nil

		return nil
	}

	if s.Message == nil {
		s.Message = &AggregateAndProof{Version: s.Version}
	}

	return s.Message.FromView(msgView)
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (s *SignedAggregateAndProof) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := s.viewType()
	if err != nil {
		return err
	}

	h, ok := any(s).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("SignedAggregateAndProof: generated SSZ code missing")
	}

	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("SignedAggregateAndProof: no view hasher for version %d", s.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (s *SignedAggregateAndProof) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return s.MarshalSSZDyn(ds, make([]byte, 0, s.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (s *SignedAggregateAndProof) MarshalSSZTo(dst []byte) ([]byte, error) {
	return s.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (s *SignedAggregateAndProof) UnmarshalSSZ(buf []byte) error {
	return s.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (s *SignedAggregateAndProof) SizeSSZ() int {
	return s.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (s *SignedAggregateAndProof) HashTreeRoot() (phase0.Root, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(s)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (s *SignedAggregateAndProof) HashTreeRootWith(hh sszutils.HashWalker) error {
	return s.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}

// MarshalJSON delegates to the per-fork SignedAggregateAndProof that matches Version.
func (s *SignedAggregateAndProof) MarshalJSON() ([]byte, error) {
	return marshalAsView(s)
}

// UnmarshalJSON delegates to the per-fork SignedAggregateAndProof that matches Version.
// Caller must set Version before calling.
func (s *SignedAggregateAndProof) UnmarshalJSON(data []byte) error {
	if err := unmarshalAsView(s, data); err != nil {
		return err
	}

	s.populateVersion(s.Version)

	return nil
}

// MarshalYAML delegates to the per-fork SignedAggregateAndProof that matches Version.
func (s *SignedAggregateAndProof) MarshalYAML() ([]byte, error) {
	return marshalAsViewYAML(s)
}

// UnmarshalYAML delegates to the per-fork SignedAggregateAndProof that matches Version.
// Caller must set Version before calling.
func (s *SignedAggregateAndProof) UnmarshalYAML(data []byte) error {
	if err := unmarshalAsViewYAML(s, data); err != nil {
		return err
	}

	s.populateVersion(s.Version)

	return nil
}
