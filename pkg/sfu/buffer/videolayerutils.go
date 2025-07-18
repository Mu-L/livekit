// Copyright 2023 LiveKit, Inc.
//
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

package buffer

import (
	"slices"

	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/logger"
)

const (
	quarterResolutionQ = "q"
	halfResolutionH    = "h"
	fullResolutionF    = "f"

	quarterResolution0 = "0"
	halfResolution1    = "1"
	fullResolution2    = "2"
)

type VideoLayersRid [DefaultMaxLayerSpatial + 1]string

var (
	videoLayersRidQHF     = VideoLayersRid{quarterResolutionQ, halfResolutionH, fullResolutionF}
	videoLayersRid012     = VideoLayersRid{quarterResolution0, halfResolution1, fullResolution2}
	DefaultVideoLayersRid = videoLayersRidQHF
)

// SIMULCAST-CODEC-TODO: these need to be codec mime aware if and when each codec suppports different layers
func LayerPresenceFromTrackInfo(trackInfo *livekit.TrackInfo) *[livekit.VideoQuality_HIGH + 1]bool {
	if trackInfo == nil || len(trackInfo.Layers) == 0 {
		return nil
	}

	var layerPresence [livekit.VideoQuality_HIGH + 1]bool
	for _, layer := range trackInfo.Layers {
		// WARNING: comparing protobuf enum
		if layer.Quality <= livekit.VideoQuality_HIGH {
			layerPresence[layer.Quality] = true
		} else {
			logger.Warnw("unexpected quality in track info", nil, "trackID", trackInfo.Sid, "trackInfo", logger.Proto(trackInfo))
		}
	}

	return &layerPresence
}

func RidToSpatialLayer(rid string, trackInfo *livekit.TrackInfo, ridSpace VideoLayersRid) int32 {
	lp := LayerPresenceFromTrackInfo(trackInfo)
	if lp == nil {
		switch rid {
		case quarterResolutionQ:
			return 0
		case halfResolutionH:
			return 1
		case fullResolutionF:
			return 2
		default:
			return 0
		}
	}

	switch rid {
	case ridSpace[0]:
		switch {
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			return 0

		default:
			// only one quality published, could be any
			return 0
		}

	case ridSpace[1]:
		switch {
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			return 1

		default:
			// only one quality published, could be any
			return 0
		}

	case ridSpace[2]:
		switch {
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			return 2

		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM]:
			logger.Warnw("unexpected rid with only two qualities, low and medium", nil, "trackID", trackInfo.Sid, "trackInfo", logger.Proto(trackInfo), "rid", ridSpace[2])
			return 1
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_HIGH]:
			logger.Warnw("unexpected rid with only two qualities, low and high", nil, "trackID", trackInfo.Sid, "trackInfo", logger.Proto(trackInfo), "rid", ridSpace[2])
			return 1
		case lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			logger.Warnw("unexpected rid with only two qualities, medium and high", nil, "trackID", trackInfo.Sid, "trackInfo", logger.Proto(trackInfo), "rid", ridSpace[2])
			return 1

		default:
			// only one quality published, could be any
			return 0
		}

	default:
		// no rid, should be single layer
		return 0
	}
}

func SpatialLayerToRid(layer int32, trackInfo *livekit.TrackInfo, ridSpace VideoLayersRid) string {
	lp := LayerPresenceFromTrackInfo(trackInfo)
	if lp == nil {
		switch layer {
		case 0:
			return quarterResolutionQ
		case 1:
			return halfResolutionH
		case 2:
			return fullResolutionF
		default:
			return quarterResolutionQ
		}
	}

	switch layer {
	case 0:
		switch {
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			return ridSpace[0]

		default:
			return ridSpace[0]
		}

	case 1:
		switch {
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			return ridSpace[1]

		default:
			return ridSpace[0]
		}

	case 2:
		switch {
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			return ridSpace[2]

		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM]:
			logger.Warnw("unexpected layer 2 with only two qualities, low and medium", nil, "trackID", trackInfo.Sid, "trackInfo", logger.Proto(trackInfo))
			return ridSpace[1]
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_HIGH]:
			logger.Warnw("unexpected layer 2 with only two qualities, low and high", nil, "trackID", trackInfo.Sid, "trackInfo", logger.Proto(trackInfo))
			return ridSpace[1]
		case lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			logger.Warnw("unexpected layer 2 with only two qualities, medium and high", nil, "trackID", trackInfo.Sid, "trackInfo", logger.Proto(trackInfo))
			return ridSpace[1]

		default:
			return ridSpace[0]
		}

	default:
		return ridSpace[0]
	}
}

func VideoQualityToRid(quality livekit.VideoQuality, trackInfo *livekit.TrackInfo, ridSpace VideoLayersRid) string {
	return SpatialLayerToRid(VideoQualityToSpatialLayer(quality, trackInfo), trackInfo, ridSpace)
}

func SpatialLayerToVideoQuality(layer int32, trackInfo *livekit.TrackInfo) livekit.VideoQuality {
	lp := LayerPresenceFromTrackInfo(trackInfo)
	if lp == nil {
		switch layer {
		case 0:
			return livekit.VideoQuality_LOW
		case 1:
			return livekit.VideoQuality_MEDIUM
		case 2:
			return livekit.VideoQuality_HIGH
		default:
			return livekit.VideoQuality_OFF
		}
	}

	switch layer {
	case 0:
		switch {
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_LOW]:
			return livekit.VideoQuality_LOW

		case lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_MEDIUM]:
			return livekit.VideoQuality_MEDIUM

		default:
			return livekit.VideoQuality_HIGH
		}

	case 1:
		switch {
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM]:
			return livekit.VideoQuality_MEDIUM

		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			return livekit.VideoQuality_HIGH

		default:
			logger.Errorw("invalid layer", nil, "trackID", trackInfo.Sid, "layer", layer, "trackInfo", logger.Proto(trackInfo))
			return livekit.VideoQuality_HIGH
		}

	case 2:
		switch {
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			return livekit.VideoQuality_HIGH

		default:
			logger.Errorw("invalid layer", nil, "trackID", trackInfo.Sid, "layer", layer, "trackInfo", logger.Proto(trackInfo))
			return livekit.VideoQuality_HIGH
		}
	}

	return livekit.VideoQuality_OFF
}

func VideoQualityToSpatialLayer(quality livekit.VideoQuality, trackInfo *livekit.TrackInfo) int32 {
	lp := LayerPresenceFromTrackInfo(trackInfo)
	if lp == nil {
		switch quality {
		case livekit.VideoQuality_LOW:
			return 0
		case livekit.VideoQuality_MEDIUM:
			return 1
		case livekit.VideoQuality_HIGH:
			return 2
		default:
			return InvalidLayerSpatial
		}
	}

	switch quality {
	case livekit.VideoQuality_LOW:
		switch {
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		default: // only one quality published, could be any
			return 0
		}

	case livekit.VideoQuality_MEDIUM:
		switch {
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_HIGH]:
			return 1

		case lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			return 0

		default: // only one quality published, could be any
			return 0
		}

	case livekit.VideoQuality_HIGH:
		switch {
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			return 2

		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_MEDIUM]:
			fallthrough
		case lp[livekit.VideoQuality_LOW] && lp[livekit.VideoQuality_HIGH]:
			fallthrough
		case lp[livekit.VideoQuality_MEDIUM] && lp[livekit.VideoQuality_HIGH]:
			return 1

		default: // only one quality published, could be any
			return 0
		}
	}

	return InvalidLayerSpatial
}

// SIMULCAST-CODEC-TODO: these need to be codec mime aware if and when each codec suppports different layers
func GetSpatialLayerForRid(rid string, ti *livekit.TrackInfo) int32 {
	if rid == "" {
		// single layer without RID
		return 0
	}

	if ti == nil {
		return InvalidLayerSpatial
	}

	for _, layer := range ti.Layers {
		if layer.Rid == rid {
			return layer.SpatialLayer
		}
	}

	if len(ti.Layers) != 0 {
		// RID present in codec, but may not be specified via signalling
		// (happens with older browsers setting a rid for SVC codecs)
		hasRid := false
		for _, layer := range ti.Layers {
			if layer.Rid != "" {
				hasRid = true
				break
			}
		}
		if !hasRid {
			return 0
		}
	}

	// SIMULCAST-CODEC-TODO - ideally should return invalid, but there are
	// VP9 publish using rid = f, if there are only two layers
	// in TrackInfo, that will be q;h and f will become invalid.
	//
	// Actually, there should be no rids for VP9 in SDP and hence
	// the above check should take effect. However, as simulcast
	// codec does not update SDP, the default rids are use when
	// vp9 (primary codec) is published and that bypasses the above
	// check.
	//
	// The full proper sequence would be
	// 1. For primary codec using SVC, there will be no rids.
	//    The above check should take effect and it should
	//    return 0 even if some publisher uses a rid like `f`.
	// 2. When secondary codec is published, update rids in the
	//    codec section in `TrackInfo`. This is a bit tricky
	//    for a couple of cases
	//    a. Browsers like Firefox use a different CID everytime.
	//       So, it cannot be matched between `AddTrack` and SDP.
	//       One option is to look for a published track with
	//       back up codec and apply it there. But, that becomes
	//       a challenge if there are multiple published tracks
	//       with pending back up codec.
	//    b. The back up codec publish SDP will have the full
	//       codec list. It should be okay to assume that the
	//       codec that will be published is the back up codec,
	//       but just something to be aware of..
	// 3. Use this function with proper mime so that proper
	//    codec section can be looked up in `TrackInfo`.
	// return InvalidLayerSpatial
	logger.Infow(
		"invalid layer for rid, returning default",
		"trackID", ti.Sid,
		"rid", rid,
		"trackInfo", logger.Proto(ti),
	)
	return 0
}

func GetSpatialLayerForVideoQuality(quality livekit.VideoQuality, ti *livekit.TrackInfo) int32 {
	if ti == nil || quality == livekit.VideoQuality_OFF {
		return InvalidLayerSpatial
	}

	for _, layer := range ti.Layers {
		if layer.Quality == quality {
			return layer.SpatialLayer
		}
	}

	if len(ti.Layers) == 0 {
		// single layer
		return 0
	}

	// requested quality is higher than available layers, return the highest available layer
	return ti.Layers[len(ti.Layers)-1].SpatialLayer
}

func GetVideoQualityForSpatialLayer(spatialLayer int32, ti *livekit.TrackInfo) livekit.VideoQuality {
	if spatialLayer == InvalidLayerSpatial || ti == nil {
		return livekit.VideoQuality_OFF
	}

	for _, layer := range ti.Layers {
		if layer.SpatialLayer == spatialLayer {
			return layer.Quality
		}
	}

	return livekit.VideoQuality_OFF
}

func isVideoLayersRidKnown(rids VideoLayersRid, knownRids VideoLayersRid) bool {
	for _, rid := range rids {
		if rid == "" {
			continue
		}

		if !slices.Contains(knownRids[:], rid) {
			return false
		}
	}

	return true
}

func NormalizeVideoLayersRid(rids VideoLayersRid) VideoLayersRid {
	out := rids

	normalize := func(knownRids VideoLayersRid) {
		idx := 0
		for _, known := range knownRids {
			if slices.Contains(rids[:], known) {
				out[idx] = known
				idx++
			}
		}
	}

	if isVideoLayersRidKnown(rids, videoLayersRidQHF) {
		normalize(videoLayersRidQHF)
	}

	if isVideoLayersRidKnown(rids, videoLayersRid012) {
		normalize(videoLayersRid012)
	}

	return out
}
