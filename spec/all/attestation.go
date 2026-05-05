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

	bitfield "github.com/OffchainLabs/go-bitfield"
	"github.com/ethpandaops/go-eth2-client/spec/electra"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/pk910/dynamic-ssz/sszutils"
)

// Attestation is a fork-agnostic attestation containing the union of fields
// from every fork. CommitteeBits is populated from Electra onwards.
type Attestation struct {
	Version         version.DataVersion
	AggregationBits bitfield.Bitlist
	Data            *phase0.AttestationData
	Signature       phase0.BLSSignature
	CommitteeBits   bitfield.Bitvector64
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (a *Attestation) viewType() (any, error) {
	switch a.Version {
	case version.DataVersionPhase0,
		version.DataVersionAltair,
		version.DataVersionBellatrix,
		version.DataVersionCapella,
		version.DataVersionDeneb:
		return (*phase0.Attestation)(nil), nil
	case version.DataVersionElectra,
		version.DataVersionFulu,
		version.DataVersionGloas,
		version.DataVersionHeze:
		return (*electra.Attestation)(nil), nil
	default:
		return nil, fmt.Errorf("Attestation: unsupported version %d", a.Version)
	}
}

// MarshalSSZDyn marshals the attestation using the view that matches Version.
func (a *Attestation) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := a.viewType()
	if err != nil {
		return nil, err
	}

	m, ok := any(a).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("Attestation: generated SSZ code missing")
	}

	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("Attestation: no view marshaler for version %d", a.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the attestation for the active Version.
func (a *Attestation) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
	view, err := a.viewType()
	if err != nil {
		return 0
	}

	s, ok := any(a).(sszutils.DynamicViewSizer)
	if !ok {
		return 0
	}

	fn := s.SizeSSZDynView(view)
	if fn == nil {
		return 0
	}

	return fn(ds)
}

// UnmarshalSSZDyn decodes the attestation into the view that matches Version.
func (a *Attestation) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := a.viewType()
	if err != nil {
		return err
	}

	u, ok := any(a).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("Attestation: generated SSZ code missing")
	}

	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("Attestation: no view unmarshaler for version %d", a.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	a.populateVersion(a.Version)

	return nil
}

// populateVersion sets Version. Attestation has no nested versionable children.
func (a *Attestation) populateVersion(v version.DataVersion) {
	a.Version = v
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (a *Attestation) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := a.viewType()
	if err != nil {
		return err
	}

	h, ok := any(a).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("Attestation: generated SSZ code missing")
	}

	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("Attestation: no view hasher for version %d", a.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (a *Attestation) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return a.MarshalSSZDyn(ds, make([]byte, 0, a.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (a *Attestation) MarshalSSZTo(dst []byte) ([]byte, error) {
	return a.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (a *Attestation) UnmarshalSSZ(buf []byte) error {
	return a.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (a *Attestation) SizeSSZ() int {
	return a.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (a *Attestation) HashTreeRoot() (phase0.Root, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(a)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (a *Attestation) HashTreeRootWith(hh sszutils.HashWalker) error {
	return a.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}
