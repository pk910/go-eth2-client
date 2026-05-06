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

	"github.com/ethpandaops/go-eth2-client/spec/altair"
	"github.com/ethpandaops/go-eth2-client/spec/bellatrix"
	"github.com/ethpandaops/go-eth2-client/spec/capella"
	"github.com/ethpandaops/go-eth2-client/spec/deneb"
	"github.com/ethpandaops/go-eth2-client/spec/electra"
	"github.com/ethpandaops/go-eth2-client/spec/gloas"
	"github.com/ethpandaops/go-eth2-client/spec/heze"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/pk910/dynamic-ssz/sszutils"
)

// SignedBeaconBlock is a fork-agnostic signed beacon block.
type SignedBeaconBlock struct {
	Version   version.DataVersion
	Message   *BeaconBlock
	Signature phase0.BLSSignature
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (s *SignedBeaconBlock) viewType() (any, error) {
	switch s.Version {
	case version.DataVersionPhase0:
		return (*phase0.SignedBeaconBlock)(nil), nil
	case version.DataVersionAltair:
		return (*altair.SignedBeaconBlock)(nil), nil
	case version.DataVersionBellatrix:
		return (*bellatrix.SignedBeaconBlock)(nil), nil
	case version.DataVersionCapella:
		return (*capella.SignedBeaconBlock)(nil), nil
	case version.DataVersionDeneb:
		return (*deneb.SignedBeaconBlock)(nil), nil
	case version.DataVersionElectra,
		version.DataVersionFulu:
		// Fulu reuses the Electra signed block schema unchanged.
		return (*electra.SignedBeaconBlock)(nil), nil
	case version.DataVersionGloas:
		return (*gloas.SignedBeaconBlock)(nil), nil
	case version.DataVersionHeze:
		return (*heze.SignedBeaconBlock)(nil), nil
	default:
		return nil, fmt.Errorf("SignedBeaconBlock: unsupported version %d", s.Version)
	}
}

// MarshalSSZDyn marshals the signed block using the view that matches Version.
func (s *SignedBeaconBlock) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := s.viewType()
	if err != nil {
		return nil, err
	}

	m, ok := any(s).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("SignedBeaconBlock: generated SSZ code missing")
	}

	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("SignedBeaconBlock: no view marshaler for version %d", s.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the signed block for the active Version.
func (s *SignedBeaconBlock) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
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

// UnmarshalSSZDyn decodes the signed block into the view that matches Version.
func (s *SignedBeaconBlock) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := s.viewType()
	if err != nil {
		return err
	}

	u, ok := any(s).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("SignedBeaconBlock: generated SSZ code missing")
	}

	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("SignedBeaconBlock: no view unmarshaler for version %d", s.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	s.populateVersion(s.Version)

	return nil
}

// populateVersion sets Version and propagates it to any nested versionable
// children allocated by the SSZ unmarshal.
func (s *SignedBeaconBlock) populateVersion(v version.DataVersion) {
	s.Version = v

	if s.Message != nil {
		s.Message.populateVersion(v)
	}
}

// ToView returns a fresh fork-specific SignedBeaconBlock populated with s's
// fields, recursing into Message via its ToView.
func (s *SignedBeaconBlock) ToView() (any, error) {
	var msg any

	var err error

	if s.Message != nil {
		msg, err = s.Message.ToView()
		if err != nil {
			return nil, err
		}
	}

	switch s.Version {
	case version.DataVersionPhase0:
		m, err := assertView[*phase0.BeaconBlock](msg, "SignedBeaconBlock.Message")
		if err != nil {
			return nil, err
		}

		return &phase0.SignedBeaconBlock{Message: m, Signature: s.Signature}, nil
	case version.DataVersionAltair:
		m, err := assertView[*altair.BeaconBlock](msg, "SignedBeaconBlock.Message")
		if err != nil {
			return nil, err
		}

		return &altair.SignedBeaconBlock{Message: m, Signature: s.Signature}, nil
	case version.DataVersionBellatrix:
		m, err := assertView[*bellatrix.BeaconBlock](msg, "SignedBeaconBlock.Message")
		if err != nil {
			return nil, err
		}

		return &bellatrix.SignedBeaconBlock{Message: m, Signature: s.Signature}, nil
	case version.DataVersionCapella:
		m, err := assertView[*capella.BeaconBlock](msg, "SignedBeaconBlock.Message")
		if err != nil {
			return nil, err
		}

		return &capella.SignedBeaconBlock{Message: m, Signature: s.Signature}, nil
	case version.DataVersionDeneb:
		m, err := assertView[*deneb.BeaconBlock](msg, "SignedBeaconBlock.Message")
		if err != nil {
			return nil, err
		}

		return &deneb.SignedBeaconBlock{Message: m, Signature: s.Signature}, nil
	case version.DataVersionElectra,
		version.DataVersionFulu:
		// Fulu reuses the Electra signed-block schema unchanged.
		m, err := assertView[*electra.BeaconBlock](msg, "SignedBeaconBlock.Message")
		if err != nil {
			return nil, err
		}

		return &electra.SignedBeaconBlock{Message: m, Signature: s.Signature}, nil
	case version.DataVersionGloas:
		m, err := assertView[*gloas.BeaconBlock](msg, "SignedBeaconBlock.Message")
		if err != nil {
			return nil, err
		}

		return &gloas.SignedBeaconBlock{Message: m, Signature: s.Signature}, nil
	case version.DataVersionHeze:
		m, err := assertView[*heze.BeaconBlock](msg, "SignedBeaconBlock.Message")
		if err != nil {
			return nil, err
		}

		return &heze.SignedBeaconBlock{Message: m, Signature: s.Signature}, nil
	default:
		return nil, fmt.Errorf("SignedBeaconBlock: unsupported version %d", s.Version)
	}
}

// FromView populates s from a fork-specific SignedBeaconBlock.
func (s *SignedBeaconBlock) FromView(view any) error {
	var msgView any

	switch v := view.(type) {
	case *phase0.SignedBeaconBlock:
		if s.Version == version.DataVersionUnknown {
			s.Version = version.DataVersionPhase0
		}

		s.Signature, msgView = v.Signature, v.Message
	case *altair.SignedBeaconBlock:
		if s.Version == version.DataVersionUnknown {
			s.Version = version.DataVersionAltair
		}

		s.Signature, msgView = v.Signature, v.Message
	case *bellatrix.SignedBeaconBlock:
		if s.Version == version.DataVersionUnknown {
			s.Version = version.DataVersionBellatrix
		}

		s.Signature, msgView = v.Signature, v.Message
	case *capella.SignedBeaconBlock:
		if s.Version == version.DataVersionUnknown {
			s.Version = version.DataVersionCapella
		}

		s.Signature, msgView = v.Signature, v.Message
	case *deneb.SignedBeaconBlock:
		if s.Version == version.DataVersionUnknown {
			s.Version = version.DataVersionDeneb
		}

		s.Signature, msgView = v.Signature, v.Message
	case *electra.SignedBeaconBlock:
		if s.Version == version.DataVersionUnknown {
			s.Version = version.DataVersionElectra
		}

		s.Signature, msgView = v.Signature, v.Message
	case *gloas.SignedBeaconBlock:
		if s.Version == version.DataVersionUnknown {
			s.Version = version.DataVersionGloas
		}

		s.Signature, msgView = v.Signature, v.Message
	case *heze.SignedBeaconBlock:
		if s.Version == version.DataVersionUnknown {
			s.Version = version.DataVersionHeze
		}

		s.Signature, msgView = v.Signature, v.Message
	default:
		return fmt.Errorf("SignedBeaconBlock: unsupported view type %T", view)
	}

	if msgView == nil {
		s.Message = nil

		return nil
	}

	if s.Message == nil {
		s.Message = &BeaconBlock{Version: s.Version}
	}

	return s.Message.FromView(msgView)
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (s *SignedBeaconBlock) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := s.viewType()
	if err != nil {
		return err
	}

	h, ok := any(s).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("SignedBeaconBlock: generated SSZ code missing")
	}

	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("SignedBeaconBlock: no view hasher for version %d", s.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (s *SignedBeaconBlock) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return s.MarshalSSZDyn(ds, make([]byte, 0, s.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (s *SignedBeaconBlock) MarshalSSZTo(dst []byte) ([]byte, error) {
	return s.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (s *SignedBeaconBlock) UnmarshalSSZ(buf []byte) error {
	return s.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (s *SignedBeaconBlock) SizeSSZ() int {
	return s.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (s *SignedBeaconBlock) HashTreeRoot() ([32]byte, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(s)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (s *SignedBeaconBlock) HashTreeRootWith(hh sszutils.HashWalker) error {
	return s.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}

// MarshalJSON delegates to the per-fork SignedBeaconBlock that matches Version.
func (s *SignedBeaconBlock) MarshalJSON() ([]byte, error) {
	return marshalAsView(s)
}

// UnmarshalJSON delegates to the per-fork SignedBeaconBlock that matches Version.
// Caller must set Version before calling.
func (s *SignedBeaconBlock) UnmarshalJSON(data []byte) error {
	if err := unmarshalAsView(s, data); err != nil {
		return err
	}

	s.populateVersion(s.Version)

	return nil
}

// MarshalYAML delegates to the per-fork SignedBeaconBlock that matches Version.
func (s *SignedBeaconBlock) MarshalYAML() ([]byte, error) {
	return marshalAsViewYAML(s)
}

// UnmarshalYAML delegates to the per-fork SignedBeaconBlock that matches Version.
// Caller must set Version before calling.
func (s *SignedBeaconBlock) UnmarshalYAML(data []byte) error {
	if err := unmarshalAsViewYAML(s, data); err != nil {
		return err
	}

	s.populateVersion(s.Version)

	return nil
}
