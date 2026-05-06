package all_test

import (
	"reflect"
	"testing"

	"github.com/ethpandaops/go-eth2-client/spec/all"
	"github.com/ethpandaops/go-eth2-client/spec/altair"
	"github.com/ethpandaops/go-eth2-client/spec/bellatrix"
	"github.com/ethpandaops/go-eth2-client/spec/capella"
	"github.com/ethpandaops/go-eth2-client/spec/deneb"
	"github.com/ethpandaops/go-eth2-client/spec/electra"
	"github.com/ethpandaops/go-eth2-client/spec/fulu"
	"github.com/ethpandaops/go-eth2-client/spec/gloas"
	"github.com/ethpandaops/go-eth2-client/spec/heze"
	"github.com/ethpandaops/go-eth2-client/spec/phase0"
	dynssz "github.com/pk910/dynamic-ssz"
)

func TestBeaconStateViews(t *testing.T) {
	ds := dynssz.GetGlobalDynSsz()
	if err := ds.ValidateType(reflect.TypeOf((*all.BeaconState)(nil)), dynssz.WithViewDescriptor((*phase0.BeaconState)(nil))); err != nil {
		t.Fatalf("Failed to validate BeaconState with phase0 view: %v", err)
	}
	if err := ds.ValidateType(reflect.TypeOf((*all.BeaconState)(nil)), dynssz.WithViewDescriptor((*altair.BeaconState)(nil))); err != nil {
		t.Fatalf("Failed to validate BeaconState with altair view: %v", err)
	}
	if err := ds.ValidateType(reflect.TypeOf((*all.BeaconState)(nil)), dynssz.WithViewDescriptor((*bellatrix.BeaconState)(nil))); err != nil {
		t.Fatalf("Failed to validate BeaconState with bellatrix view: %v", err)
	}
	if err := ds.ValidateType(reflect.TypeOf((*all.BeaconState)(nil)), dynssz.WithViewDescriptor((*capella.BeaconState)(nil))); err != nil {
		t.Fatalf("Failed to validate BeaconState with capella view: %v", err)
	}
	if err := ds.ValidateType(reflect.TypeOf((*all.BeaconState)(nil)), dynssz.WithViewDescriptor((*deneb.BeaconState)(nil))); err != nil {
		t.Fatalf("Failed to validate BeaconState with deneb view: %v", err)
	}
	if err := ds.ValidateType(reflect.TypeOf((*all.BeaconState)(nil)), dynssz.WithViewDescriptor((*electra.BeaconState)(nil))); err != nil {
		t.Fatalf("Failed to validate BeaconState with electra view: %v", err)
	}
	if err := ds.ValidateType(reflect.TypeOf((*all.BeaconState)(nil)), dynssz.WithViewDescriptor((*fulu.BeaconState)(nil))); err != nil {
		t.Fatalf("Failed to validate BeaconState with fulu view: %v", err)
	}
	if err := ds.ValidateType(reflect.TypeOf((*all.BeaconState)(nil)), dynssz.WithViewDescriptor((*gloas.BeaconState)(nil))); err != nil {
		t.Fatalf("Failed to validate BeaconState with gloas view: %v", err)
	}
	if err := ds.ValidateType(reflect.TypeOf((*all.BeaconState)(nil)), dynssz.WithViewDescriptor((*heze.BeaconState)(nil))); err != nil {
		t.Fatalf("Failed to validate BeaconState with heze view: %v", err)
	}
}
