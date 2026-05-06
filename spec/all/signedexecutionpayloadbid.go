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

	"github.com/ethpandaops/go-eth2-client/spec/gloas"
	"github.com/ethpandaops/go-eth2-client/spec/heze"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/pk910/dynamic-ssz/sszutils"
)

// SignedExecutionPayloadBid is a fork-agnostic signed execution payload bid.
type SignedExecutionPayloadBid struct {
	Version   version.DataVersion
	Message   *ExecutionPayloadBid
	Signature phase0.BLSSignature
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (s *SignedExecutionPayloadBid) viewType() (any, error) {
	switch s.Version {
	case version.DataVersionGloas:
		return (*gloas.SignedExecutionPayloadBid)(nil), nil
	case version.DataVersionHeze:
		return (*heze.SignedExecutionPayloadBid)(nil), nil
	default:
		return nil, fmt.Errorf("SignedExecutionPayloadBid: unsupported version %d", s.Version)
	}
}

// MarshalSSZDyn marshals the signed bid using the view that matches Version.
func (s *SignedExecutionPayloadBid) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := s.viewType()
	if err != nil {
		return nil, err
	}

	m, ok := any(s).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("SignedExecutionPayloadBid: generated SSZ code missing")
	}

	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("SignedExecutionPayloadBid: no view marshaler for version %d", s.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the signed bid for the active Version.
func (s *SignedExecutionPayloadBid) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
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

// UnmarshalSSZDyn decodes the signed bid into the view that matches Version.
func (s *SignedExecutionPayloadBid) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := s.viewType()
	if err != nil {
		return err
	}

	u, ok := any(s).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("SignedExecutionPayloadBid: generated SSZ code missing")
	}

	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("SignedExecutionPayloadBid: no view unmarshaler for version %d", s.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	s.populateVersion(s.Version)

	return nil
}

// populateVersion sets Version and propagates it to the inner message.
func (s *SignedExecutionPayloadBid) populateVersion(v version.DataVersion) {
	s.Version = v

	if s.Message != nil {
		s.Message.populateVersion(v)
	}
}

// ToView returns a fresh fork-specific SignedExecutionPayloadBid populated
// with s's fields, recursing into Message via its ToView.
func (s *SignedExecutionPayloadBid) ToView() (any, error) {
	var msg any

	var err error

	if s.Message != nil {
		msg, err = s.Message.ToView()
		if err != nil {
			return nil, err
		}
	}

	switch s.Version {
	case version.DataVersionGloas:
		gm, err := assertView[*gloas.ExecutionPayloadBid](msg, "SignedExecutionPayloadBid.Message")
		if err != nil {
			return nil, err
		}

		return &gloas.SignedExecutionPayloadBid{Message: gm, Signature: s.Signature}, nil
	case version.DataVersionHeze:
		hm, err := assertView[*heze.ExecutionPayloadBid](msg, "SignedExecutionPayloadBid.Message")
		if err != nil {
			return nil, err
		}

		return &heze.SignedExecutionPayloadBid{Message: hm, Signature: s.Signature}, nil
	default:
		return nil, fmt.Errorf("SignedExecutionPayloadBid: unsupported version %d", s.Version)
	}
}

// FromView populates s from a fork-specific SignedExecutionPayloadBid.
func (s *SignedExecutionPayloadBid) FromView(view any) error {
	var msgView any

	switch v := view.(type) {
	case *gloas.SignedExecutionPayloadBid:
		if s.Version == version.DataVersionUnknown {
			s.Version = version.DataVersionGloas
		}

		s.Signature = v.Signature

		if v.Message != nil {
			msgView = v.Message
		}
	case *heze.SignedExecutionPayloadBid:
		if s.Version == version.DataVersionUnknown {
			s.Version = version.DataVersionHeze
		}

		s.Signature = v.Signature

		if v.Message != nil {
			msgView = v.Message
		}
	default:
		return fmt.Errorf("SignedExecutionPayloadBid: unsupported view type %T", view)
	}

	if msgView == nil {
		s.Message = nil

		return nil
	}

	if s.Message == nil {
		s.Message = &ExecutionPayloadBid{Version: s.Version}
	}

	return s.Message.FromView(msgView)
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (s *SignedExecutionPayloadBid) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := s.viewType()
	if err != nil {
		return err
	}

	h, ok := any(s).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("SignedExecutionPayloadBid: generated SSZ code missing")
	}

	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("SignedExecutionPayloadBid: no view hasher for version %d", s.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (s *SignedExecutionPayloadBid) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return s.MarshalSSZDyn(ds, make([]byte, 0, s.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (s *SignedExecutionPayloadBid) MarshalSSZTo(dst []byte) ([]byte, error) {
	return s.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (s *SignedExecutionPayloadBid) UnmarshalSSZ(buf []byte) error {
	return s.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (s *SignedExecutionPayloadBid) SizeSSZ() int {
	return s.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (s *SignedExecutionPayloadBid) HashTreeRoot() (phase0.Root, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(s)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (s *SignedExecutionPayloadBid) HashTreeRootWith(hh sszutils.HashWalker) error {
	return s.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}

// MarshalJSON delegates to the per-fork SignedExecutionPayloadBid that matches Version.
func (s *SignedExecutionPayloadBid) MarshalJSON() ([]byte, error) {
	return marshalAsView(s)
}

// UnmarshalJSON delegates to the per-fork SignedExecutionPayloadBid that matches Version.
// Caller must set Version before calling.
func (s *SignedExecutionPayloadBid) UnmarshalJSON(data []byte) error {
	if err := unmarshalAsView(s, data); err != nil {
		return err
	}

	s.populateVersion(s.Version)

	return nil
}

// MarshalYAML delegates to the per-fork SignedExecutionPayloadBid that matches Version.
func (s *SignedExecutionPayloadBid) MarshalYAML() ([]byte, error) {
	return marshalAsViewYAML(s)
}

// UnmarshalYAML delegates to the per-fork SignedExecutionPayloadBid that matches Version.
// Caller must set Version before calling.
func (s *SignedExecutionPayloadBid) UnmarshalYAML(data []byte) error {
	if err := unmarshalAsViewYAML(s, data); err != nil {
		return err
	}

	s.populateVersion(s.Version)

	return nil
}
