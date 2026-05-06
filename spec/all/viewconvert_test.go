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

package all_test

import (
	"reflect"
	"testing"

	"github.com/ethpandaops/go-eth2-client/spec/all"
	"github.com/ethpandaops/go-eth2-client/spec/electra"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/go-eth2-client/spec/version"
)

// TestToViewFromViewLeaf round-trips a leaf type (Attestation) through ToView
// and FromView for both phase0 and electra views.
func TestToViewFromViewLeaf(t *testing.T) {
	t.Run("phase0", func(t *testing.T) {
		a := &all.Attestation{
			Version:         version.DataVersionDeneb,
			AggregationBits: []byte{0x01, 0x02},
			Data:            &phase0.AttestationData{Slot: 7},
			Signature:       phase0.BLSSignature{0xab, 0xcd},
		}

		view, err := a.ToView()
		if err != nil {
			t.Fatalf("ToView: %v", err)
		}

		pa, ok := view.(*phase0.Attestation)
		if !ok {
			t.Fatalf("expected *phase0.Attestation, got %T", view)
		}

		if pa.Data.Slot != 7 {
			t.Fatalf("Slot lost: %d", pa.Data.Slot)
		}

		// Round-trip through FromView.
		rt := &all.Attestation{Version: version.DataVersionDeneb}
		if err := rt.FromView(pa); err != nil {
			t.Fatalf("FromView: %v", err)
		}

		if rt.Data.Slot != 7 {
			t.Fatalf("RT data slot: %d", rt.Data.Slot)
		}
	})

	t.Run("electra adds CommitteeBits", func(t *testing.T) {
		a := &all.Attestation{
			Version:         version.DataVersionElectra,
			AggregationBits: []byte{0x10},
			Data:            &phase0.AttestationData{Slot: 9},
			CommitteeBits:   make([]byte, 8),
		}
		a.CommitteeBits[0] = 0x42

		view, err := a.ToView()
		if err != nil {
			t.Fatalf("ToView: %v", err)
		}

		ea, ok := view.(*electra.Attestation)
		if !ok {
			t.Fatalf("expected *electra.Attestation, got %T", view)
		}

		if ea.CommitteeBits[0] != 0x42 {
			t.Fatalf("CommitteeBits lost: %x", ea.CommitteeBits[0])
		}

		rt := &all.Attestation{Version: version.DataVersionElectra}
		if err := rt.FromView(ea); err != nil {
			t.Fatalf("FromView: %v", err)
		}

		if rt.CommitteeBits[0] != 0x42 {
			t.Fatalf("RT CommitteeBits: %x", rt.CommitteeBits[0])
		}
	})
}

// TestToViewFromViewNested round-trips an AttesterSlashing whose children are
// nested versionable IndexedAttestations.
func TestToViewFromViewNested(t *testing.T) {
	asrc := &all.AttesterSlashing{
		Version: version.DataVersionPhase0,
		Attestation1: &all.IndexedAttestation{
			Version:          version.DataVersionPhase0,
			AttestingIndices: []uint64{1, 2, 3},
			Data:             &phase0.AttestationData{Slot: 11},
			Signature:        phase0.BLSSignature{0xaa},
		},
		Attestation2: &all.IndexedAttestation{
			Version:          version.DataVersionPhase0,
			AttestingIndices: []uint64{4, 5},
			Data:             &phase0.AttestationData{Slot: 12},
			Signature:        phase0.BLSSignature{0xbb},
		},
	}

	view, err := asrc.ToView()
	if err != nil {
		t.Fatalf("ToView: %v", err)
	}

	pasl, ok := view.(*phase0.AttesterSlashing)
	if !ok {
		t.Fatalf("expected *phase0.AttesterSlashing, got %T", view)
	}

	if pasl.Attestation1 == nil || pasl.Attestation2 == nil {
		t.Fatal("nested attestations not converted")
	}

	if pasl.Attestation1.Data.Slot != 11 || pasl.Attestation2.Data.Slot != 12 {
		t.Fatal("nested data lost")
	}

	rt := &all.AttesterSlashing{Version: version.DataVersionPhase0}
	if err := rt.FromView(pasl); err != nil {
		t.Fatalf("FromView: %v", err)
	}

	if rt.Attestation1 == nil || rt.Attestation2 == nil {
		t.Fatal("RT nested missing")
	}

	if rt.Attestation1.Data.Slot != 11 || rt.Attestation2.Data.Slot != 12 {
		t.Fatal("RT data lost")
	}

	if rt.Attestation1.Version != version.DataVersionPhase0 {
		t.Fatalf("nested Version not set: %d", rt.Attestation1.Version)
	}

	if !reflect.DeepEqual(rt.Attestation1.AttestingIndices, asrc.Attestation1.AttestingIndices) {
		t.Fatal("AttestingIndices lost")
	}
}

// TestToViewFromViewBeaconBlock exercises a deeper tree (SignedBeaconBlock →
// BeaconBlock → BeaconBlockBody → AttesterSlashings → IndexedAttestation).
func TestToViewFromViewBeaconBlock(t *testing.T) {
	src := &all.SignedBeaconBlock{
		Version:   version.DataVersionPhase0,
		Signature: phase0.BLSSignature{0xff},
		Message: &all.BeaconBlock{
			Version:       version.DataVersionPhase0,
			Slot:          42,
			ProposerIndex: 7,
			Body: &all.BeaconBlockBody{
				Version:           version.DataVersionPhase0,
				ETH1Data:          &phase0.ETH1Data{DepositCount: 5},
				ProposerSlashings: []*phase0.ProposerSlashing{},
				AttesterSlashings: []*all.AttesterSlashing{},
				Attestations:      []*all.Attestation{},
				Deposits:          []*phase0.Deposit{},
				VoluntaryExits:    []*phase0.SignedVoluntaryExit{},
			},
		},
	}

	view, err := src.ToView()
	if err != nil {
		t.Fatalf("ToView: %v", err)
	}

	psb, ok := view.(*phase0.SignedBeaconBlock)
	if !ok {
		t.Fatalf("expected *phase0.SignedBeaconBlock, got %T", view)
	}

	if psb.Message.Slot != 42 || psb.Message.ProposerIndex != 7 {
		t.Fatalf("BeaconBlock fields lost: %+v", psb.Message)
	}

	if psb.Message.Body == nil || psb.Message.Body.ETH1Data.DepositCount != 5 {
		t.Fatal("Body fields lost")
	}

	rt := &all.SignedBeaconBlock{Version: version.DataVersionPhase0}
	if err := rt.FromView(psb); err != nil {
		t.Fatalf("FromView: %v", err)
	}

	if rt.Message == nil || rt.Message.Body == nil {
		t.Fatal("nested missing after FromView")
	}

	if rt.Message.Slot != 42 || rt.Message.Body.ETH1Data.DepositCount != 5 {
		t.Fatal("RT fields lost")
	}

	// Version must propagate to all nested versionables.
	if rt.Message.Version != version.DataVersionPhase0 ||
		rt.Message.Body.Version != version.DataVersionPhase0 {
		t.Fatalf("nested Version not propagated: msg=%d body=%d",
			rt.Message.Version, rt.Message.Body.Version)
	}
}

// TestToViewUnsupportedVersion verifies ToView returns an error when Version
// is unset.
func TestToViewUnsupportedVersion(t *testing.T) {
	a := &all.Attestation{}
	if _, err := a.ToView(); err == nil {
		t.Fatal("expected error for unset Version")
	}
}

// TestFromViewUnsupportedType verifies FromView returns an error when the view
// type doesn't match.
func TestFromViewUnsupportedType(t *testing.T) {
	a := &all.Attestation{}
	if err := a.FromView("not a view"); err == nil {
		t.Fatal("expected error for invalid view type")
	}
}
