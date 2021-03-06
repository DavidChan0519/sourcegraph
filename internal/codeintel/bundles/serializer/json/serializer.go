package json

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/serializer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

type jsonSerializer struct{}

var _ serializer.Serializer = &jsonSerializer{}

func New() serializer.Serializer {
	return &jsonSerializer{}
}

func (*jsonSerializer) MarshalDocumentData(d types.DocumentData) ([]byte, error) {
	rangePairs := make([]interface{}, 0, len(d.Ranges))
	for k, v := range d.Ranges {
		if v.MonikerIDs == nil {
			v.MonikerIDs = []types.ID{}
		}

		vs := SerializingRange{
			StartLine:          v.StartLine,
			StartCharacter:     v.StartCharacter,
			EndLine:            v.EndLine,
			EndCharacter:       v.EndCharacter,
			DefinitionResultID: v.DefinitionResultID,
			ReferenceResultID:  v.ReferenceResultID,
			HoverResultID:      v.HoverResultID,
			MonikerIDs:         SerializingTaggedValue{Type: "set", Value: v.MonikerIDs},
		}

		rangePairs = append(rangePairs, []interface{}{k, vs})
	}

	hoverResultPairs := make([]interface{}, 0, len(d.HoverResults))
	for k, v := range d.HoverResults {
		hoverResultPairs = append(hoverResultPairs, []interface{}{k, v})
	}

	monikerPairs := make([]interface{}, 0, len(d.Monikers))
	for k, v := range d.Monikers {
		monikerPairs = append(monikerPairs, []interface{}{k, v})
	}

	packageInformationPairs := make([]interface{}, 0, len(d.PackageInformation))
	for k, v := range d.PackageInformation {
		packageInformationPairs = append(packageInformationPairs, []interface{}{k, v})
	}

	encoded, err := json.Marshal(SerializingDocument{
		Ranges:             SerializingTaggedValue{Type: "map", Value: rangePairs},
		HoverResults:       SerializingTaggedValue{Type: "map", Value: hoverResultPairs},
		Monikers:           SerializingTaggedValue{Type: "map", Value: monikerPairs},
		PackageInformation: SerializingTaggedValue{Type: "map", Value: packageInformationPairs},
	})
	if err != nil {
		return nil, err
	}

	return compress(encoded)
}

func (jsonSerializer) MarshalResultChunkData(rc types.ResultChunkData) ([]byte, error) {
	documentPathPairs := make([]interface{}, 0, len(rc.DocumentPaths))
	for k, v := range rc.DocumentPaths {
		documentPathPairs = append(documentPathPairs, []interface{}{k, v})
	}

	documentIDRangeIDPairs := make([]interface{}, 0, len(rc.DocumentIDRangeIDs))
	for k, v := range rc.DocumentIDRangeIDs {
		documentIDRangeIDPairs = append(documentIDRangeIDPairs, []interface{}{k, v})
	}

	encoded, err := json.Marshal(SerializingResultChunk{
		DocumentPaths:      SerializingTaggedValue{Type: "map", Value: documentPathPairs},
		DocumentIDRangeIDs: SerializingTaggedValue{Type: "map", Value: documentIDRangeIDPairs},
	})
	if err != nil {
		return nil, err
	}

	return compress(encoded)
}

func (jsonSerializer) UnmarshalDocumentData(data []byte) (types.DocumentData, error) {
	var payload SerializedDocument
	if err := unmarshalGzippedJSON(data, &payload); err != nil {
		return types.DocumentData{}, err
	}

	ranges, err := unmarshalWrappedRanges(payload.Ranges.Value)
	if err != nil {
		return types.DocumentData{}, err
	}

	hoverResults, err := unmarshalWrappedHoverResults(payload.HoverResults.Value)
	if err != nil {
		return types.DocumentData{}, err
	}

	monikers, err := unmarshalWrappedMonikers(payload.Monikers.Value)
	if err != nil {
		return types.DocumentData{}, err
	}

	packageInformation, err := unmarshalWrappedPackageInformation(payload.PackageInformation.Value)
	if err != nil {
		return types.DocumentData{}, err
	}

	return types.DocumentData{
		Ranges:             ranges,
		HoverResults:       hoverResults,
		Monikers:           monikers,
		PackageInformation: packageInformation,
	}, nil
}

func (jsonSerializer) UnmarshalResultChunkData(data []byte) (types.ResultChunkData, error) {
	var payload SerializedResultChunk
	if err := unmarshalGzippedJSON(data, &payload); err != nil {
		return types.ResultChunkData{}, err
	}

	documentPaths, err := unmarshalWrappedDocumentPaths(payload.DocumentPaths.Value)
	if err != nil {
		return types.ResultChunkData{}, err
	}

	documentIDRangeIDs, err := unmarshalWrappedDocumentIDRangeIDs(payload.DocumentIDRangeIDs.Value)
	if err != nil {
		return types.ResultChunkData{}, err
	}

	return types.ResultChunkData{
		DocumentPaths:      documentPaths,
		DocumentIDRangeIDs: documentIDRangeIDs,
	}, nil
}

func unmarshalWrappedRanges(pairs []json.RawMessage) (map[types.ID]types.RangeData, error) {
	m := map[types.ID]types.RangeData{}
	for _, pair := range pairs {
		var id ID
		var value SerializedRange

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		var monikerIDs []types.ID
		for _, value := range value.MonikerIDs.Value {
			var id ID
			if err := json.Unmarshal([]byte(value), &id); err != nil {
				return nil, err
			}

			monikerIDs = append(monikerIDs, types.ID(id))
		}

		m[types.ID(id)] = types.RangeData{
			StartLine:          value.StartLine,
			StartCharacter:     value.StartCharacter,
			EndLine:            value.EndLine,
			EndCharacter:       value.EndCharacter,
			DefinitionResultID: types.ID(value.DefinitionResultID),
			ReferenceResultID:  types.ID(value.ReferenceResultID),
			HoverResultID:      types.ID(value.HoverResultID),
			MonikerIDs:         monikerIDs,
		}
	}

	return m, nil
}

func unmarshalWrappedHoverResults(pairs []json.RawMessage) (map[types.ID]string, error) {
	m := map[types.ID]string{}
	for _, pair := range pairs {
		var id ID
		var value string

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[types.ID(id)] = value
	}

	return m, nil
}

func unmarshalWrappedMonikers(pairs []json.RawMessage) (map[types.ID]types.MonikerData, error) {
	m := map[types.ID]types.MonikerData{}
	for _, pair := range pairs {
		var id ID
		var value SerializedMoniker

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[types.ID(id)] = types.MonikerData{
			Kind:                 value.Kind,
			Scheme:               value.Scheme,
			Identifier:           value.Identifier,
			PackageInformationID: types.ID(value.PackageInformationID),
		}
	}

	return m, nil
}

func unmarshalWrappedPackageInformation(pairs []json.RawMessage) (map[types.ID]types.PackageInformationData, error) {
	m := map[types.ID]types.PackageInformationData{}
	for _, pair := range pairs {
		var id ID
		var value SerializedPackageInformation

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[types.ID(id)] = types.PackageInformationData{
			Name:    value.Name,
			Version: value.Version,
		}
	}

	return m, nil
}

func unmarshalWrappedDocumentPaths(pairs []json.RawMessage) (map[types.ID]string, error) {
	m := map[types.ID]string{}
	for _, pair := range pairs {
		var id ID
		var value string

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[types.ID(id)] = value
	}

	return m, nil
}

func unmarshalWrappedDocumentIDRangeIDs(pairs []json.RawMessage) (map[types.ID][]types.DocumentIDRangeID, error) {
	m := map[types.ID][]types.DocumentIDRangeID{}
	for _, pair := range pairs {
		var id ID
		var value []SerializedDocumentIDRangeID

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		var documentIDRangeIDs []types.DocumentIDRangeID
		for _, v := range value {
			documentIDRangeIDs = append(documentIDRangeIDs, types.DocumentIDRangeID{
				DocumentID: types.ID(v.DocumentID),
				RangeID:    types.ID(v.RangeID),
			})
		}

		m[types.ID(id)] = documentIDRangeIDs
	}

	return m, nil
}
