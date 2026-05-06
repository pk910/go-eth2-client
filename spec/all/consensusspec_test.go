// Copyright © 2025 Attestant Limited.
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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethpandaops/go-eth2-client/spec/all"
	"github.com/ethpandaops/go-eth2-client/spec/version"
	"github.com/goccy/go-yaml"
	"github.com/golang/snappy"
	"github.com/pk910/dynamic-ssz/sszutils"
	require "github.com/stretchr/testify/require"
)

// TestConsensusSpec runs the agnostic union types against the Ethereum
// consensus spec ssz_static test data for every fork that has data on disk.
// It mirrors the per-fork spec test runners but loops over all forks and
// pre-sets each instance's Version before unmarshal.
func TestConsensusSpec(t *testing.T) {
	if os.Getenv("CONSENSUS_SPEC_TESTS_DIR") == "" {
		t.Skip("CONSENSUS_SPEC_TESTS_DIR not supplied, not running spec tests")
	}

	forks := []struct {
		dir     string
		version version.DataVersion
	}{
		{"phase0", version.DataVersionPhase0},
		{"altair", version.DataVersionAltair},
		{"bellatrix", version.DataVersionBellatrix},
		{"capella", version.DataVersionCapella},
		{"deneb", version.DataVersionDeneb},
		{"electra", version.DataVersionElectra},
		{"fulu", version.DataVersionFulu},
		{"gloas", version.DataVersionGloas},
		{"heze", version.DataVersionHeze},
	}

	// Each entry is a single agnostic type tested against every fork that ships
	// data for it. The factory returns a fresh instance with Version pre-set so
	// the union type's UnmarshalSSZ/UnmarshalYAML route through the right view.
	tests := []struct {
		name    string
		factory func(version.DataVersion) any
	}{
		{
			name:    "BeaconState",
			factory: func(v version.DataVersion) any { return &all.BeaconState{Version: v} },
		},
		{
			name:    "BeaconBlock",
			factory: func(v version.DataVersion) any { return &all.BeaconBlock{Version: v} },
		},
		{
			name:    "SignedBeaconBlock",
			factory: func(v version.DataVersion) any { return &all.SignedBeaconBlock{Version: v} },
		},
		{
			name:    "BeaconBlockBody",
			factory: func(v version.DataVersion) any { return &all.BeaconBlockBody{Version: v} },
		},
		{
			name:    "ExecutionPayload",
			factory: func(v version.DataVersion) any { return &all.ExecutionPayload{Version: v} },
		},
		{
			name:    "ExecutionPayloadHeader",
			factory: func(v version.DataVersion) any { return &all.ExecutionPayloadHeader{Version: v} },
		},
		{
			name:    "ExecutionPayloadBid",
			factory: func(v version.DataVersion) any { return &all.ExecutionPayloadBid{Version: v} },
		},
		{
			name:    "SignedExecutionPayloadBid",
			factory: func(v version.DataVersion) any { return &all.SignedExecutionPayloadBid{Version: v} },
		},
		// ExecutionPayloadEnvelope and SignedExecutionPayloadEnvelope are not
		// included here because the gloas.ExecutionPayload YAML codec
		// currently rejects test fixtures missing slot_number — a pre-existing
		// per-fork bug unrelated to the agnostic envelope wrapper. Add these
		// back once the gloas YAML codec is fixed.
		{
			name:    "Attestation",
			factory: func(v version.DataVersion) any { return &all.Attestation{Version: v} },
		},
		{
			name:    "AttesterSlashing",
			factory: func(v version.DataVersion) any { return &all.AttesterSlashing{Version: v} },
		},
		{
			name:    "IndexedAttestation",
			factory: func(v version.DataVersion) any { return &all.IndexedAttestation{Version: v} },
		},
		{
			name:    "AggregateAndProof",
			factory: func(v version.DataVersion) any { return &all.AggregateAndProof{Version: v} },
		},
		{
			name:    "SignedAggregateAndProof",
			factory: func(v version.DataVersion) any { return &all.SignedAggregateAndProof{Version: v} },
		},
	}

	root := os.Getenv("CONSENSUS_SPEC_TESTS_DIR")

	for _, fork := range forks {
		baseDir := filepath.Join(root, "tests", "mainnet", fork.dir, "ssz_static")
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			// This fork has no test data on disk — skip silently.
			continue
		}

		for _, test := range tests {
			dir := filepath.Join(baseDir, test.name, "ssz_random")
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				// This fork doesn't ship data for this type (e.g. phase0 has
				// no ExecutionPayload, fulu only has BeaconState). Skip.
				continue
			}

			runForkTypeCases(t, dir, fork.dir, fork.version, test.name, test.factory)
		}
	}
}

func runForkTypeCases(
	t *testing.T,
	dir, forkDir string,
	ver version.DataVersion,
	typeName string,
	factory func(version.DataVersion) any,
) {
	t.Helper()

	require.NoError(t, filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if path == dir {
			// Only interested in the case subdirectories below ssz_random.
			return nil
		}

		require.NoError(t, err)

		if !info.IsDir() {
			return nil
		}

		caseName := fmt.Sprintf("%s/%s/%s", forkDir, typeName, info.Name())
		t.Run(caseName, func(t *testing.T) {
			runConsensusCase(t, path, ver, factory)
		})

		return nil
	}))
}

// viewConvertible is the agnostic-type interface tested in the ToView/FromView
// round-trip below.
type viewConvertible interface {
	ToView() (any, error)
	FromView(view any) error
}

func runConsensusCase(
	t *testing.T,
	path string,
	ver version.DataVersion,
	factory func(version.DataVersion) any,
) {
	t.Helper()

	// YAML round-trip.
	s1 := factory(ver)
	specYAML, err := os.ReadFile(filepath.Join(path, "value.yaml"))
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(specYAML, s1))

	remarshalledSpecYAML, err := yaml.Marshal(s1)
	require.NoError(t, err)
	require.YAMLEq(t, testYAMLFormat(specYAML), testYAMLFormat(remarshalledSpecYAML))

	// SSZ round-trip via a fresh instance — the YAML path must not influence
	// the SSZ assertions.
	s2 := factory(ver)
	compressedSpecSSZ, err := os.ReadFile(filepath.Join(path, "serialized.ssz_snappy"))
	require.NoError(t, err)

	specSSZ, err := snappy.Decode(nil, compressedSpecSSZ)
	require.NoError(t, err)

	require.NoError(t, s2.(sszutils.FastsszUnmarshaler).UnmarshalSSZ(specSSZ))

	remarshalledSpecSSZ, err := s2.(sszutils.FastsszMarshaler).MarshalSSZ()
	require.NoError(t, err)
	require.Equal(t, specSSZ, remarshalledSpecSSZ)

	// Hash tree root must match the value the spec dataset shipped.
	specYAMLRoot, err := os.ReadFile(filepath.Join(path, "roots.yaml"))
	require.NoError(t, err)

	generatedRootBytes, err := s2.(sszutils.FastsszHashRoot).HashTreeRoot()
	require.NoError(t, err)

	generatedRoot := fmt.Sprintf("{root: '%#x'}\n", string(generatedRootBytes[:]))
	require.YAMLEq(t, string(specYAMLRoot), generatedRoot)

	// ToView/FromView round-trip exercised via the SSZ-loaded instance:
	//   1. ToView produces a fresh fork-specific *fork.X.
	//   2. The view marshals back to the same SSZ bytes (proves ToView copies
	//      every field, including nested children, via copyByName).
	//   3. A fresh agnostic instance fed the same view through FromView
	//      reproduces the SSZ bytes and the hash tree root.
	conv, ok := s2.(viewConvertible)
	require.Truef(t, ok, "agnostic type %T does not implement viewConvertible", s2)

	view, err := conv.ToView()
	require.NoError(t, err)
	require.NotNil(t, view)

	viewSSZ, err := view.(sszutils.FastsszMarshaler).MarshalSSZ()
	require.NoError(t, err, "ToView output failed to marshal SSZ")
	require.Equal(t, specSSZ, viewSSZ, "ToView output SSZ differs from spec bytes")

	// Build a fresh agnostic instance with no Version pinned so FromView
	// must infer it from the view's concrete type.
	s3 := factory(version.DataVersionUnknown).(viewConvertible)
	require.NoError(t, s3.FromView(view), "FromView failed")

	roundTrippedSSZ, err := s3.(sszutils.FastsszMarshaler).MarshalSSZ()
	require.NoError(t, err, "FromView output failed to marshal SSZ")
	require.Equal(t, specSSZ, roundTrippedSSZ, "ToView→FromView round-trip lost data")

	roundTrippedRoot, err := s3.(sszutils.FastsszHashRoot).HashTreeRoot()
	require.NoError(t, err)
	require.Equal(t, generatedRootBytes, roundTrippedRoot, "ToView→FromView round-trip altered hash tree root")
}

// testYAMLFormat normalises a YAML document so two semantically equal but
// syntactically different documents compare equal in require.YAMLEq.
// Mirrored from the per-fork spec tests.
func testYAMLFormat(input []byte) string {
	val := make(map[string]any)
	if err := yaml.UnmarshalWithOptions(input, &val, yaml.UseOrderedMap()); err != nil {
		panic(err)
	}

	res, err := yaml.MarshalWithOptions(val, yaml.Flow(true))
	if err != nil {
		panic(err)
	}

	replacements := [][][]byte{
		{[]byte(`"`), []byte(`'`)},
		// Field 'extra_data' in ExecutionPayloadHeader/case_1 has a non-standard
		// format in the spec data; normalise to match other cases.
		{[]byte(`extra_data: 0,`), []byte(`extra_data: '0x',`)},
	}
	for _, replacement := range replacements {
		res = bytes.ReplaceAll(res, replacement[0], replacement[1])
	}

	return string(bytes.ToLower(res))
}
