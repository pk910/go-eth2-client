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
	"github.com/ethpandaops/go-eth2-client/spec/capella"
	"github.com/ethpandaops/go-eth2-client/spec/deneb"
	"github.com/ethpandaops/go-eth2-client/spec/gloas"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
	"github.com/holiman/uint256"
	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/pk910/dynamic-ssz/sszutils"
)

// ExecutionPayload is a fork-agnostic execution payload containing the union
// of fields from every fork. Fields populated on a given instance depend on
// Version.
type ExecutionPayload struct {
	Version         version.DataVersion
	ParentHash      phase0.Hash32
	FeeRecipient    bellatrix.ExecutionAddress
	StateRoot       phase0.Root
	ReceiptsRoot    phase0.Root
	LogsBloom       [256]byte
	PrevRandao      [32]byte
	BlockNumber     uint64
	GasLimit        uint64
	GasUsed         uint64
	Timestamp       uint64
	ExtraData       []byte
	BaseFeePerGasLE [32]byte
	BaseFeePerGas   *uint256.Int
	BlockHash       phase0.Hash32
	Transactions    []bellatrix.Transaction
	Withdrawals     []*capella.Withdrawal
	BlobGasUsed     uint64
	ExcessBlobGas   uint64
	BlockAccessList gloas.BlockAccessList
	SlotNumber      uint64
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (e *ExecutionPayload) viewType() (any, error) {
	switch e.Version {
	case version.DataVersionBellatrix:
		return (*bellatrix.ExecutionPayload)(nil), nil
	case version.DataVersionCapella:
		return (*capella.ExecutionPayload)(nil), nil
	case version.DataVersionDeneb:
		return (*deneb.ExecutionPayload)(nil), nil
	case version.DataVersionGloas:
		return (*gloas.ExecutionPayload)(nil), nil
	default:
		return nil, fmt.Errorf("ExecutionPayload: unsupported version %d", e.Version)
	}
}

// MarshalSSZDyn marshals the payload using the view that matches Version.
func (e *ExecutionPayload) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := e.viewType()
	if err != nil {
		return nil, err
	}

	m, ok := any(e).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("ExecutionPayload: generated SSZ code missing")
	}

	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("ExecutionPayload: no view marshaler for version %d", e.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the payload for the active Version.
func (e *ExecutionPayload) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
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

// UnmarshalSSZDyn decodes the payload into the view that matches Version.
func (e *ExecutionPayload) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := e.viewType()
	if err != nil {
		return err
	}

	u, ok := any(e).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("ExecutionPayload: generated SSZ code missing")
	}

	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("ExecutionPayload: no view unmarshaler for version %d", e.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	e.populateVersion(e.Version)

	return nil
}

// populateVersion sets Version. ExecutionPayload has no nested versionable
// children — its fields are primitives, fork-specific types, or simple slices.
func (e *ExecutionPayload) populateVersion(v version.DataVersion) {
	e.Version = v
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (e *ExecutionPayload) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := e.viewType()
	if err != nil {
		return err
	}

	h, ok := any(e).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("ExecutionPayload: generated SSZ code missing")
	}

	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("ExecutionPayload: no view hasher for version %d", e.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (e *ExecutionPayload) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return e.MarshalSSZDyn(ds, make([]byte, 0, e.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (e *ExecutionPayload) MarshalSSZTo(dst []byte) ([]byte, error) {
	return e.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (e *ExecutionPayload) UnmarshalSSZ(buf []byte) error {
	return e.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (e *ExecutionPayload) SizeSSZ() int {
	return e.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (e *ExecutionPayload) HashTreeRoot() (phase0.Root, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(e)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (e *ExecutionPayload) HashTreeRootWith(hh sszutils.HashWalker) error {
	return e.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}

// MarshalJSON delegates to the per-fork ExecutionPayload that matches Version.
func (e *ExecutionPayload) MarshalJSON() ([]byte, error) {
	return marshalAsView(e)
}

// UnmarshalJSON delegates to the per-fork ExecutionPayload that matches Version.
// Caller must set Version before calling.
func (e *ExecutionPayload) UnmarshalJSON(data []byte) error {
	if err := unmarshalAsView(e, data); err != nil {
		return err
	}

	e.populateVersion(e.Version)

	return nil
}

// MarshalYAML delegates to the per-fork ExecutionPayload that matches Version.
func (e *ExecutionPayload) MarshalYAML() ([]byte, error) {
	return marshalAsViewYAML(e)
}

// UnmarshalYAML delegates to the per-fork ExecutionPayload that matches Version.
// Caller must set Version before calling.
func (e *ExecutionPayload) UnmarshalYAML(data []byte) error {
	if err := unmarshalAsViewYAML(e, data); err != nil {
		return err
	}

	e.populateVersion(e.Version)

	return nil
}
