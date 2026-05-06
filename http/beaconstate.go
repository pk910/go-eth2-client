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

// BeaconState fetches a beacon state.
func (s *Service) BeaconState(ctx context.Context,
	opts *api.BeaconStateOpts,
) (
	*api.Response[*spec.VersionedBeaconState],
	error,
) {
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

	httpResponse, err := s.get(ctx, endpoint, "", &opts.Common, true)
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

func (s *Service) beaconStateFromSSZ(ctx context.Context, res *httpResponse) (*api.Response[*spec.VersionedBeaconState], error) {
	response := &api.Response[*spec.VersionedBeaconState]{
		Data: &spec.VersionedBeaconState{
			Version: res.consensusVersion,
		},
		Metadata: metadataFromHeaders(res.headers),
	}

	var dynSSZ *dynssz.DynSsz

	if s.customSpecSupport {
		specs, err := s.Spec(ctx, &api.SpecOpts{})
		if err != nil {
			return nil, errors.Join(errors.New("failed to request specs"), err)
		}

		dynSSZ = dynssz.NewDynSsz(specs.Data)
	}

	var err error

	switch res.consensusVersion {
	case spec.DataVersionPhase0:
		response.Data.Phase0 = &phase0.BeaconState{}
		if s.customSpecSupport {
			err = dynSSZ.UnmarshalSSZ(response.Data.Phase0, res.body)
		} else {
			err = response.Data.Phase0.UnmarshalSSZ(res.body)
		}

		if err != nil {
			return nil, errors.Join(errors.New("failed to decode phase0 beacon state"), err)
		}
	case spec.DataVersionAltair:
		response.Data.Altair = &altair.BeaconState{}
		if s.customSpecSupport {
			err = dynSSZ.UnmarshalSSZ(response.Data.Altair, res.body)
		} else {
			err = response.Data.Altair.UnmarshalSSZ(res.body)
		}

		if err != nil {
			return nil, errors.Join(errors.New("failed to decode altair beacon state"), err)
		}
	case spec.DataVersionBellatrix:
		response.Data.Bellatrix = &bellatrix.BeaconState{}
		if s.customSpecSupport {
			err = dynSSZ.UnmarshalSSZ(response.Data.Bellatrix, res.body)
		} else {
			err = response.Data.Bellatrix.UnmarshalSSZ(res.body)
		}

		if err != nil {
			return nil, errors.Join(errors.New("failed to decode bellatrix beacon state"), err)
		}
	case spec.DataVersionCapella:
		response.Data.Capella = &capella.BeaconState{}
		if s.customSpecSupport {
			err = dynSSZ.UnmarshalSSZ(response.Data.Capella, res.body)
		} else {
			err = response.Data.Capella.UnmarshalSSZ(res.body)
		}

		if err != nil {
			return nil, errors.Join(errors.New("failed to decode capella beacon state"), err)
		}
	case spec.DataVersionDeneb:
		response.Data.Deneb = &deneb.BeaconState{}
		if s.customSpecSupport {
			err = dynSSZ.UnmarshalSSZ(response.Data.Deneb, res.body)
		} else {
			err = response.Data.Deneb.UnmarshalSSZ(res.body)
		}

		if err != nil {
			return nil, errors.Join(errors.New("failed to decode deneb beacon state"), err)
		}
	case spec.DataVersionElectra:
		response.Data.Electra = &electra.BeaconState{}
		if s.customSpecSupport {
			err = dynSSZ.UnmarshalSSZ(response.Data.Electra, res.body)
		} else {
			err = response.Data.Electra.UnmarshalSSZ(res.body)
		}

		if err != nil {
			return nil, errors.Join(errors.New("failed to decode electra beacon state"), err)
		}
	case spec.DataVersionFulu:
		response.Data.Fulu = &fulu.BeaconState{}
		if s.customSpecSupport {
			err = dynSSZ.UnmarshalSSZ(response.Data.Fulu, res.body)
		} else {
			err = response.Data.Fulu.UnmarshalSSZ(res.body)
		}

		if err != nil {
			return nil, errors.Join(errors.New("failed to decode fulu beacon state"), err)
		}
	case spec.DataVersionGloas:
		response.Data.Gloas = &gloas.BeaconState{}
		if s.customSpecSupport {
			err = dynSSZ.UnmarshalSSZ(response.Data.Gloas, res.body)
		} else {
			err = response.Data.Gloas.UnmarshalSSZ(res.body)
		}
		if err != nil {
			return nil, errors.Join(errors.New("failed to decode gloas beacon state"), err)
		}
	case spec.DataVersionHeze:
		response.Data.Heze = &heze.BeaconState{}
		if s.customSpecSupport {
			err = dynSSZ.UnmarshalSSZ(response.Data.Heze, res.body)
		} else {
			err = response.Data.Heze.UnmarshalSSZ(res.body)
		}
		if err != nil {
			return nil, errors.Join(errors.New("failed to decode heze beacon state"), err)
		}
	default:
		return nil, fmt.Errorf("unhandled state version %s", res.consensusVersion)
	}

	return response, nil
}

// AgnosticBeaconState fetches a beacon state and decodes it directly into a
// fork-agnostic *all.BeaconState. The Version is set from the consensus
// version header before unmarshaling so the union type's view-aware codec
// dispatches into the correct fork's schema without any intermediate
// fork-specific allocation.
func (s *Service) AgnosticBeaconState(ctx context.Context,
	opts *api.BeaconStateOpts,
) (
	*api.Response[*all.BeaconState],
	error,
) {
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

	httpResponse, err := s.get(ctx, endpoint, "", &opts.Common, true)
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

// dynSSZForRequest returns a dynssz instance to use for SSZ codecs. When
// customSpecSupport is enabled it returns a per-request instance built from
// the connected node's spec values; otherwise it returns the global instance.
func (s *Service) dynSSZForRequest(ctx context.Context) (*dynssz.DynSsz, error) {
	if !s.customSpecSupport {
		return dynssz.GetGlobalDynSsz(), nil
	}

	specs, err := s.Spec(ctx, &api.SpecOpts{})
	if err != nil {
		return nil, errors.Join(errors.New("failed to request specs"), err)
	}

	return dynssz.NewDynSsz(specs.Data), nil
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
