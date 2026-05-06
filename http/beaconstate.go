// Copyright © 2020 - 2024 Attestant Limited.
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

package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	client "github.com/ethpandaops/go-eth2-client"
	"github.com/ethpandaops/go-eth2-client/api"
	"github.com/ethpandaops/go-eth2-client/spec"
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

// BeaconState fetches a beacon state given a state ID and decodes it directly
// into the per-fork view stored on a *spec.VersionedBeaconState. No
// intermediate copy.
func (s *Service) BeaconState(ctx context.Context,
	opts *api.BeaconStateOpts,
) (
	*api.Response[*spec.VersionedBeaconState],
	error,
) {
	httpResponse, err := s.fetchBeaconState(ctx, opts)
	if err != nil {
		return nil, err
	}

	switch httpResponse.contentType {
	case ContentTypeSSZ:
		return s.beaconStateFromSSZ(ctx, httpResponse)
	case ContentTypeJSON:
		return s.beaconStateFromJSON(httpResponse)
	default:
		return nil, fmt.Errorf("unhandled content type %v", httpResponse.contentType)
	}
}

// AgnosticBeaconState fetches a beacon state and decodes it directly into a
// fork-agnostic *all.BeaconState. The Version is set from the consensus
// version header before unmarshaling so the union type's view-aware codec
// dispatches into the correct fork's schema. No intermediate copy.
func (s *Service) AgnosticBeaconState(ctx context.Context,
	opts *api.BeaconStateOpts,
) (
	*api.Response[*all.BeaconState],
	error,
) {
	httpResponse, err := s.fetchBeaconState(ctx, opts)
	if err != nil {
		return nil, err
	}

	state := &all.BeaconState{Version: httpResponse.consensusVersion}

	switch httpResponse.contentType {
	case ContentTypeSSZ:
		ds, err := s.dynSSZForRequest(ctx)
		if err != nil {
			return nil, err
		}

		if err := state.UnmarshalSSZDyn(ds, httpResponse.body); err != nil {
			return nil, errors.Join(fmt.Errorf("failed to decode %s beacon state", httpResponse.consensusVersion), err)
		}
	case ContentTypeJSON:
		if err := state.UnmarshalJSON(httpResponse.body); err != nil {
			return nil, errors.Join(fmt.Errorf("failed to decode %s beacon state", httpResponse.consensusVersion), err)
		}
	default:
		return nil, fmt.Errorf("unhandled content type %v", httpResponse.contentType)
	}

	return &api.Response[*all.BeaconState]{
		Data:     state,
		Metadata: metadataFromHeaders(httpResponse.headers),
	}, nil
}

// fetchBeaconState performs the GET request shared by BeaconState and
// AgnosticBeaconState: validates opts and hits the endpoint.
func (s *Service) fetchBeaconState(ctx context.Context,
	opts *api.BeaconStateOpts,
) (*httpResponse, error) {
	if err := s.assertIsActive(ctx); err != nil {
		return nil, err
	}

	if opts == nil {
		return nil, client.ErrNoOptions
	}

	if opts.State == "" {
		return nil, errors.Join(errors.New("no state specified"), client.ErrInvalidOptions)
	}

	endpoint := fmt.Sprintf("/eth/v2/debug/beacon/states/%s", opts.State)

	return s.get(ctx, endpoint, "", &opts.Common, true)
}

func (s *Service) beaconStateFromSSZ(ctx context.Context, res *httpResponse) (*api.Response[*spec.VersionedBeaconState], error) {
	response := &api.Response[*spec.VersionedBeaconState]{
		Data: &spec.VersionedBeaconState{
			Version: res.consensusVersion,
		},
		Metadata: metadataFromHeaders(res.headers),
	}

	dynSSZ, err := s.dynSSZForRequest(ctx)
	if err != nil {
		return nil, err
	}

	switch res.consensusVersion {
	case spec.DataVersionPhase0:
		response.Data.Phase0 = &phase0.BeaconState{}
		err = dynSSZ.UnmarshalSSZ(response.Data.Phase0, res.body)
	case spec.DataVersionAltair:
		response.Data.Altair = &altair.BeaconState{}
		err = dynSSZ.UnmarshalSSZ(response.Data.Altair, res.body)
	case spec.DataVersionBellatrix:
		response.Data.Bellatrix = &bellatrix.BeaconState{}
		err = dynSSZ.UnmarshalSSZ(response.Data.Bellatrix, res.body)
	case spec.DataVersionCapella:
		response.Data.Capella = &capella.BeaconState{}
		err = dynSSZ.UnmarshalSSZ(response.Data.Capella, res.body)
	case spec.DataVersionDeneb:
		response.Data.Deneb = &deneb.BeaconState{}
		err = dynSSZ.UnmarshalSSZ(response.Data.Deneb, res.body)
	case spec.DataVersionElectra:
		response.Data.Electra = &electra.BeaconState{}
		err = dynSSZ.UnmarshalSSZ(response.Data.Electra, res.body)
	case spec.DataVersionFulu:
		response.Data.Fulu = &fulu.BeaconState{}
		err = dynSSZ.UnmarshalSSZ(response.Data.Fulu, res.body)
	case spec.DataVersionGloas:
		response.Data.Gloas = &gloas.BeaconState{}
		err = dynSSZ.UnmarshalSSZ(response.Data.Gloas, res.body)
	case spec.DataVersionHeze:
		response.Data.Heze = &heze.BeaconState{}
		err = dynSSZ.UnmarshalSSZ(response.Data.Heze, res.body)
	default:
		return nil, fmt.Errorf("unhandled state version %s", res.consensusVersion)
	}

	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to decode %s beacon state", res.consensusVersion), err)
	}

	return response, nil
}

// dynSSZForRequest returns the cached dynssz instance for the current spec
// snapshot, fetching the spec lazily on the first call (and rebuilding the
// instance when clearStaticValues invalidates the cache). The instance's
// internal type cache is reused across calls, which is the whole point of
// caching it here rather than newing one up per request.
func (s *Service) dynSSZForRequest(ctx context.Context) (*dynssz.DynSsz, error) {
	if !s.customSpecSupport {
		return dynssz.GetGlobalDynSsz(), nil
	}

	s.specMutex.RLock()
	cached := s.dynSSZ
	s.specMutex.RUnlock()

	if cached != nil {
		return cached, nil
	}

	// Trigger Spec() which fetches+caches both the spec map and the dynssz
	// instance built from it.
	if _, err := s.Spec(ctx, &api.SpecOpts{}); err != nil {
		return nil, errors.Join(errors.New("failed to request specs"), err)
	}

	s.specMutex.RLock()
	defer s.specMutex.RUnlock()

	if s.dynSSZ != nil {
		return s.dynSSZ, nil
	}

	return dynssz.GetGlobalDynSsz(), nil
}

func (*Service) beaconStateFromJSON(res *httpResponse) (*api.Response[*spec.VersionedBeaconState], error) {
	response := &api.Response[*spec.VersionedBeaconState]{
		Data: &spec.VersionedBeaconState{
			Version: res.consensusVersion,
		},
	}

	var err error

	switch res.consensusVersion {
	case spec.DataVersionPhase0:
		response.Data.Phase0, response.Metadata, err = decodeJSONResponse(bytes.NewReader(res.body), &phase0.BeaconState{})
	case spec.DataVersionAltair:
		response.Data.Altair, response.Metadata, err = decodeJSONResponse(bytes.NewReader(res.body), &altair.BeaconState{})
	case spec.DataVersionBellatrix:
		response.Data.Bellatrix, response.Metadata, err = decodeJSONResponse(bytes.NewReader(res.body), &bellatrix.BeaconState{})
	case spec.DataVersionCapella:
		response.Data.Capella, response.Metadata, err = decodeJSONResponse(bytes.NewReader(res.body), &capella.BeaconState{})
	case spec.DataVersionDeneb:
		response.Data.Deneb, response.Metadata, err = decodeJSONResponse(bytes.NewReader(res.body), &deneb.BeaconState{})
	case spec.DataVersionElectra:
		response.Data.Electra, response.Metadata, err = decodeJSONResponse(bytes.NewReader(res.body), &electra.BeaconState{})
	case spec.DataVersionFulu:
		response.Data.Fulu, response.Metadata, err = decodeJSONResponse(bytes.NewReader(res.body), &fulu.BeaconState{})
	case spec.DataVersionGloas:
		response.Data.Gloas, response.Metadata, err = decodeJSONResponse(bytes.NewReader(res.body), &gloas.BeaconState{})
	case spec.DataVersionHeze:
		response.Data.Heze, response.Metadata, err = decodeJSONResponse(bytes.NewReader(res.body), &heze.BeaconState{})
	default:
		err = fmt.Errorf("unsupported version %s", res.consensusVersion)
	}

	if err != nil {
		return nil, err
	}

	return response, nil
}
