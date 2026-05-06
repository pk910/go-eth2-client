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

// AttesterSlashing is a fork-agnostic attester slashing. The wrapped
// IndexedAttestation grows in capacity from Electra onwards.
type AttesterSlashing struct {
	Version      version.DataVersion
	Attestation1 *IndexedAttestation
	Attestation2 *IndexedAttestation
}

// viewType returns the fork-specific schema type pointer used as the view
// descriptor for the active Version.
func (a *AttesterSlashing) viewType() (any, error) {
	switch a.Version {
	case version.DataVersionPhase0,
		version.DataVersionAltair,
		version.DataVersionBellatrix,
		version.DataVersionCapella,
		version.DataVersionDeneb:
		return (*phase0.AttesterSlashing)(nil), nil
	case version.DataVersionElectra,
		version.DataVersionFulu,
		version.DataVersionGloas,
		version.DataVersionHeze:
		return (*electra.AttesterSlashing)(nil), nil
	default:
		return nil, fmt.Errorf("AttesterSlashing: unsupported version %d", a.Version)
	}
}

// MarshalSSZDyn marshals the slashing using the view that matches Version.
func (a *AttesterSlashing) MarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) ([]byte, error) {
	view, err := a.viewType()
	if err != nil {
		return nil, err
	}

	m, ok := any(a).(sszutils.DynamicViewMarshaler)
	if !ok {
		return nil, errors.New("AttesterSlashing: generated SSZ code missing")
	}

	fn := m.MarshalSSZDynView(view)
	if fn == nil {
		return nil, fmt.Errorf("AttesterSlashing: no view marshaler for version %d", a.Version)
	}

	return fn(ds, buf)
}

// SizeSSZDyn returns the SSZ size of the slashing for the active Version.
func (a *AttesterSlashing) SizeSSZDyn(ds sszutils.DynamicSpecs) int {
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

// UnmarshalSSZDyn decodes the slashing into the view that matches Version.
func (a *AttesterSlashing) UnmarshalSSZDyn(ds sszutils.DynamicSpecs, buf []byte) error {
	view, err := a.viewType()
	if err != nil {
		return err
	}

	u, ok := any(a).(sszutils.DynamicViewUnmarshaler)
	if !ok {
		return errors.New("AttesterSlashing: generated SSZ code missing")
	}

	fn := u.UnmarshalSSZDynView(view)
	if fn == nil {
		return fmt.Errorf("AttesterSlashing: no view unmarshaler for version %d", a.Version)
	}

	if err := fn(ds, buf); err != nil {
		return err
	}

	a.populateVersion(a.Version)

	return nil
}

// populateVersion sets Version and propagates it to the inner indexed
// attestations.
func (a *AttesterSlashing) populateVersion(v version.DataVersion) {
	a.Version = v

	if a.Attestation1 != nil {
		a.Attestation1.populateVersion(v)
	}

	if a.Attestation2 != nil {
		a.Attestation2.populateVersion(v)
	}
}

// ToView returns a fresh fork-specific AttesterSlashing populated with a's
// fields, recursing into the inner indexed attestations via their ToView.
func (a *AttesterSlashing) ToView() (any, error) {
	var att1, att2 any

	var err error

	if a.Attestation1 != nil {
		att1, err = a.Attestation1.ToView()
		if err != nil {
			return nil, err
		}
	}

	if a.Attestation2 != nil {
		att2, err = a.Attestation2.ToView()
		if err != nil {
			return nil, err
		}
	}

	switch a.Version {
	case version.DataVersionPhase0,
		version.DataVersionAltair,
		version.DataVersionBellatrix,
		version.DataVersionCapella,
		version.DataVersionDeneb:
		pa1, err := assertView[*phase0.IndexedAttestation](att1, "AttesterSlashing.Attestation1")
		if err != nil {
			return nil, err
		}

		pa2, err := assertView[*phase0.IndexedAttestation](att2, "AttesterSlashing.Attestation2")
		if err != nil {
			return nil, err
		}

		return &phase0.AttesterSlashing{Attestation1: pa1, Attestation2: pa2}, nil
	case version.DataVersionElectra,
		version.DataVersionFulu,
		version.DataVersionGloas,
		version.DataVersionHeze:
		ea1, err := assertView[*electra.IndexedAttestation](att1, "AttesterSlashing.Attestation1")
		if err != nil {
			return nil, err
		}

		ea2, err := assertView[*electra.IndexedAttestation](att2, "AttesterSlashing.Attestation2")
		if err != nil {
			return nil, err
		}

		return &electra.AttesterSlashing{Attestation1: ea1, Attestation2: ea2}, nil
	default:
		return nil, fmt.Errorf("AttesterSlashing: unsupported version %d", a.Version)
	}
}

// FromView populates a from a fork-specific AttesterSlashing, recursing into
// the inner indexed attestations via their FromView.
func (a *AttesterSlashing) FromView(view any) error {
	switch v := view.(type) {
	case *phase0.AttesterSlashing:
		if a.Version == version.DataVersionUnknown {
			a.Version = version.DataVersionPhase0
		}

		return a.fromIndexedAttestations(v.Attestation1, v.Attestation2)
	case *electra.AttesterSlashing:
		if a.Version == version.DataVersionUnknown {
			a.Version = version.DataVersionElectra
		}

		return a.fromIndexedAttestations(v.Attestation1, v.Attestation2)
	default:
		return fmt.Errorf("AttesterSlashing: unsupported view type %T", view)
	}
}

func (a *AttesterSlashing) fromIndexedAttestations(att1, att2 any) error {
	if att1 != nil {
		if a.Attestation1 == nil {
			a.Attestation1 = &IndexedAttestation{Version: a.Version}
		}

		if err := a.Attestation1.FromView(att1); err != nil {
			return err
		}
	} else {
		a.Attestation1 = nil
	}

	if att2 != nil {
		if a.Attestation2 == nil {
			a.Attestation2 = &IndexedAttestation{Version: a.Version}
		}

		if err := a.Attestation2.FromView(att2); err != nil {
			return err
		}
	} else {
		a.Attestation2 = nil
	}

	return nil
}

// HashTreeRootWithDyn computes the SSZ hash tree root using the active Version's view.
func (a *AttesterSlashing) HashTreeRootWithDyn(ds sszutils.DynamicSpecs, hh sszutils.HashWalker) error {
	view, err := a.viewType()
	if err != nil {
		return err
	}

	h, ok := any(a).(sszutils.DynamicViewHashRoot)
	if !ok {
		return errors.New("AttesterSlashing: generated SSZ code missing")
	}

	fn := h.HashTreeRootWithDynView(view)
	if fn == nil {
		return fmt.Errorf("AttesterSlashing: no view hasher for version %d", a.Version)
	}

	return fn(ds, hh)
}

// MarshalSSZ implements the fastssz.Marshaler interface.
func (a *AttesterSlashing) MarshalSSZ() ([]byte, error) {
	ds := dynssz.GetGlobalDynSsz()

	return a.MarshalSSZDyn(ds, make([]byte, 0, a.SizeSSZDyn(ds)))
}

// MarshalSSZTo implements the fastssz.Marshaler interface.
func (a *AttesterSlashing) MarshalSSZTo(dst []byte) ([]byte, error) {
	return a.MarshalSSZDyn(dynssz.GetGlobalDynSsz(), dst)
}

// UnmarshalSSZ implements the fastssz.Unmarshaler interface.
func (a *AttesterSlashing) UnmarshalSSZ(buf []byte) error {
	return a.UnmarshalSSZDyn(dynssz.GetGlobalDynSsz(), buf)
}

// SizeSSZ implements the fastssz.Marshaler interface.
func (a *AttesterSlashing) SizeSSZ() int {
	return a.SizeSSZDyn(dynssz.GetGlobalDynSsz())
}

// HashTreeRoot implements the fastssz.HashRoot interface.
func (a *AttesterSlashing) HashTreeRoot() (phase0.Root, error) {
	return dynssz.GetGlobalDynSsz().HashTreeRoot(a)
}

// HashTreeRootWith implements the fastssz.HashRoot interface.
func (a *AttesterSlashing) HashTreeRootWith(hh sszutils.HashWalker) error {
	return a.HashTreeRootWithDyn(dynssz.GetGlobalDynSsz(), hh)
}

// MarshalJSON delegates to the per-fork AttesterSlashing that matches Version.
func (a *AttesterSlashing) MarshalJSON() ([]byte, error) {
	return marshalAsView(a)
}

// UnmarshalJSON delegates to the per-fork AttesterSlashing that matches Version.
// Caller must set Version before calling.
func (a *AttesterSlashing) UnmarshalJSON(data []byte) error {
	if err := unmarshalAsView(a, data); err != nil {
		return err
	}

	a.populateVersion(a.Version)

	return nil
}

// MarshalYAML delegates to the per-fork AttesterSlashing that matches Version.
func (a *AttesterSlashing) MarshalYAML() ([]byte, error) {
	return marshalAsViewYAML(a)
}

// UnmarshalYAML delegates to the per-fork AttesterSlashing that matches Version.
// Caller must set Version before calling.
func (a *AttesterSlashing) UnmarshalYAML(data []byte) error {
	if err := unmarshalAsViewYAML(a, data); err != nil {
		return err
	}

	a.populateVersion(a.Version)

	return nil
}
