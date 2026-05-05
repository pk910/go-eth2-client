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

// IndexedAttestation is a fork-agnostic indexed attestation. The
// AttestingIndices ssz-max grows from Electra onwards.
type IndexedAttestation struct {
	Version          version.DataVersion
	AttestingIndices []uint64
	Data             *phase0.AttestationData
	Signature        phase0.BLSSignature
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (i *IndexedAttestation) viewType() (any, error) {
	switch i.Version {
	case version.DataVersionPhase0,
		version.DataVersionAltair,
		version.DataVersionBellatrix,
		version.DataVersionCapella,
		version.DataVersionDeneb:
		return (*phase0.IndexedAttestation)(nil), nil
	case version.DataVersionElectra,
		version.DataVersionFulu,
		version.DataVersionGloas,
		version.DataVersionHeze:
		return (*electra.IndexedAttestation)(nil), nil
	default:
		return nil, fmt.Errorf("IndexedAttestation: unsupported version %d", i.Version)
	}
}

// MarshalSSZDyn marshals the indexed attestation using the view that matches Version.
func (i *IndexedAttestation) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := i.viewType()
	if err != nil {
		return nil, err
	}

	m, ok := any(i).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("IndexedAttestation: generated SSZ code missing")
	}

	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("IndexedAttestation: no view marshaler for version %d", i.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the indexed attestation for the active Version.
func (i *IndexedAttestation) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
	view, err := i.viewType()
	if err != nil {
		return 0
	}

	s, ok := any(i).(sszutils.DynamicViewSizer)
	if !ok {
		return 0
	}

	fn := s.SizeSSZDynView(view)
	if fn == nil {
		return 0
	}

	return fn(ds)
}

// UnmarshalSSZDyn decodes the indexed attestation into the view that matches Version.
func (i *IndexedAttestation) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := i.viewType()
	if err != nil {
		return err
	}

	u, ok := any(i).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("IndexedAttestation: generated SSZ code missing")
	}

	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("IndexedAttestation: no view unmarshaler for version %d", i.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	i.populateVersion(i.Version)

	return nil
}

// populateVersion sets Version. IndexedAttestation has no nested versionable
// children.
func (i *IndexedAttestation) populateVersion(v version.DataVersion) {
	i.Version = v
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (i *IndexedAttestation) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := i.viewType()
	if err != nil {
		return err
	}

	h, ok := any(i).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("IndexedAttestation: generated SSZ code missing")
	}

	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("IndexedAttestation: no view hasher for version %d", i.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (i *IndexedAttestation) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return i.MarshalSSZDyn(ds, make([]byte, 0, i.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (i *IndexedAttestation) MarshalSSZTo(dst []byte) ([]byte, error) {
	return i.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (i *IndexedAttestation) UnmarshalSSZ(buf []byte) error {
	return i.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (i *IndexedAttestation) SizeSSZ() int {
	return i.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (i *IndexedAttestation) HashTreeRoot() (phase0.Root, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(i)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (i *IndexedAttestation) HashTreeRootWith(hh sszutils.HashWalker) error {
	return i.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}
