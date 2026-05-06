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

	"github.com/ethpandaops/go-eth2-client/spec/bellatrix"
	"github.com/ethpandaops/go-eth2-client/spec/deneb"
	"github.com/ethpandaops/go-eth2-client/spec/gloas"
	"github.com/ethpandaops/go-eth2-client/spec/heze"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/pk910/dynamic-ssz/sszutils"
)

// ExecutionPayloadBid is a fork-agnostic execution payload bid (EIP-7732 and
// later) containing the union of fields from every fork. Fields populated on
// a given instance depend on Version.
type ExecutionPayloadBid struct {
	Version               version.DataVersion
	ParentBlockHash       phase0.Hash32
	ParentBlockRoot       phase0.Root
	BlockHash             phase0.Hash32
	PrevRandao            phase0.Root
	FeeRecipient          bellatrix.ExecutionAddress
	GasLimit              uint64
	BuilderIndex          gloas.BuilderIndex
	Slot                  phase0.Slot
	Value                 phase0.Gwei
	ExecutionPayment      phase0.Gwei
	BlobKZGCommitments    []deneb.KZGCommitment
	ExecutionRequestsRoot phase0.Root
	InclusionListBits     []byte
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (e *ExecutionPayloadBid) viewType() (any, error) {
	switch e.Version {
	case version.DataVersionGloas:
		return (*gloas.ExecutionPayloadBid)(nil), nil
	case version.DataVersionHeze:
		return (*heze.ExecutionPayloadBid)(nil), nil
	default:
		return nil, fmt.Errorf("ExecutionPayloadBid: unsupported version %d", e.Version)
	}
}

// MarshalSSZDyn marshals the bid using the view that matches Version.
func (e *ExecutionPayloadBid) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := e.viewType()
	if err != nil {
		return nil, err
	}

	m, ok := any(e).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("ExecutionPayloadBid: generated SSZ code missing")
	}

	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("ExecutionPayloadBid: no view marshaler for version %d", e.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the bid for the active Version.
func (e *ExecutionPayloadBid) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
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

// UnmarshalSSZDyn decodes the bid into the view that matches Version.
func (e *ExecutionPayloadBid) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := e.viewType()
	if err != nil {
		return err
	}

	u, ok := any(e).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("ExecutionPayloadBid: generated SSZ code missing")
	}

	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("ExecutionPayloadBid: no view unmarshaler for version %d", e.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	e.populateVersion(e.Version)

	return nil
}

// populateVersion sets Version. ExecutionPayloadBid has no nested versionable
// children.
func (e *ExecutionPayloadBid) populateVersion(v version.DataVersion) {
	e.Version = v
}

// ToView returns a fresh fork-specific ExecutionPayloadBid populated with e's
// ToView returns a fresh fork-specific ExecutionPayloadBid populated with e's
// fields.
func (e *ExecutionPayloadBid) ToView() (any, error) {
	return toViewByCopy(e)
}

// FromView populates e from a fork-specific ExecutionPayloadBid.
func (e *ExecutionPayloadBid) FromView(view any) error {
	v, err := executionPayloadBidVersion(view)
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

// executionPayloadBidVersion maps an ExecutionPayloadBid view type to its
// DataVersion.
func executionPayloadBidVersion(view any) (version.DataVersion, error) {
	switch view.(type) {
	case *gloas.ExecutionPayloadBid:
		return version.DataVersionGloas, nil
	case *heze.ExecutionPayloadBid:
		return version.DataVersionHeze, nil
	default:
		return version.DataVersionUnknown, fmt.Errorf("ExecutionPayloadBid: unsupported view type %T", view)
	}
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (e *ExecutionPayloadBid) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := e.viewType()
	if err != nil {
		return err
	}

	h, ok := any(e).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("ExecutionPayloadBid: generated SSZ code missing")
	}

	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("ExecutionPayloadBid: no view hasher for version %d", e.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (e *ExecutionPayloadBid) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return e.MarshalSSZDyn(ds, make([]byte, 0, e.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (e *ExecutionPayloadBid) MarshalSSZTo(dst []byte) ([]byte, error) {
	return e.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (e *ExecutionPayloadBid) UnmarshalSSZ(buf []byte) error {
	return e.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (e *ExecutionPayloadBid) SizeSSZ() int {
	return e.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (e *ExecutionPayloadBid) HashTreeRoot() ([32]byte, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(e)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (e *ExecutionPayloadBid) HashTreeRootWith(hh sszutils.HashWalker) error {
	return e.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}

// MarshalJSON delegates to the per-fork ExecutionPayloadBid that matches Version.
func (e *ExecutionPayloadBid) MarshalJSON() ([]byte, error) {
	return marshalAsView(e)
}

// UnmarshalJSON delegates to the per-fork ExecutionPayloadBid that matches Version.
// Caller must set Version before calling.
func (e *ExecutionPayloadBid) UnmarshalJSON(data []byte) error {
	if err := unmarshalAsView(e, data); err != nil {
		return err
	}

	e.populateVersion(e.Version)

	return nil
}

// MarshalYAML delegates to the per-fork ExecutionPayloadBid that matches Version.
func (e *ExecutionPayloadBid) MarshalYAML() ([]byte, error) {
	return marshalAsViewYAML(e)
}

// UnmarshalYAML delegates to the per-fork ExecutionPayloadBid that matches Version.
// Caller must set Version before calling.
func (e *ExecutionPayloadBid) UnmarshalYAML(data []byte) error {
	if err := unmarshalAsViewYAML(e, data); err != nil {
		return err
	}

	e.populateVersion(e.Version)

	return nil
}
