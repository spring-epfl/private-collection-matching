package psm

import (
	"strings"

	"github.com/ldsec/lattigo/v2/bfv"
)

type PsiType int

const (
	PSI_PSI PsiType = iota
	PSI_CA
)

var psiTypeMap = map[string]PsiType{
	"psi": PSI_PSI,
	"ca":  PSI_CA,
}

type MatchingType int

const (
	MATCHING_NONE MatchingType = iota
	MATCHING_TVERSKY
	MATCHING_TVERSKY_PLAIN
	MATCHING_FPSM
)

var matchingTypeMap = map[string]MatchingType{
	"":              MATCHING_NONE,
	"tversky":       MATCHING_TVERSKY,
	"tversky-plain": MATCHING_TVERSKY_PLAIN,
	"fpsm":          MATCHING_FPSM,
}

type AggregationType int

const (
	AGGREGATION_NONE AggregationType = iota
	AGGREGATION_X_MS
	AGGREGATION_CA_MS
)

var aggregationTypeMap = map[string]AggregationType{
	"":      AGGREGATION_NONE,
	"x-ms":  AGGREGATION_X_MS,
	"ca-ms": AGGREGATION_CA_MS,
}

type QueryType struct {
	IsSmallDomain bool
	Psi           PsiType         // [psi, psi-ca]
	Matching      MatchingType    // [fpsm, tversky, tversky-plain]
	Aggregation   AggregationType // [x-ms, ca-ms, ""]
}

func ParsePsiString(str *string) (PsiType, bool) {
	t, ok := psiTypeMap[strings.ToLower(*str)]
	return t, ok
}

func ParseMatchingString(str *string) (MatchingType, bool) {
	t, ok := matchingTypeMap[strings.ToLower(*str)]
	return t, ok
}

func ParseAggregationString(str *string) (AggregationType, bool) {
	t, ok := aggregationTypeMap[strings.ToLower(*str)]
	return t, ok
}

func NewQueryType(use_small_domain bool, psi PsiType, psm MatchingType, aggregation AggregationType) (*QueryType, error) {
	return &QueryType{use_small_domain, psi, psm, aggregation}, nil
}

type clientKey struct {
	pk  *bfv.PublicKey
	evk *bfv.EvaluationKey
}

type psiQuery struct {
	queryType     QueryType
	clientSetSize int
	ctx           *bfv.Ciphertext
}

func (key *clientKey) MarshalBinary() (data []byte, err error) {
	var buff []byte
	if data, err = key.pk.MarshalBinary(); err != nil {
		return nil, err
	}
	if buff, err = key.evk.Rlk.MarshalBinary(); err != nil {
		return nil, err
	}
	data = append(data, buff...)
	if buff, err = key.evk.Rtks.MarshalBinary(); err != nil {
		return nil, err
	}
	data = append(data, buff...)
	return data, nil
}
func (query *psiQuery) MarshalBinary() (data []byte, err error) {
	var buff []byte
	if buff, err = query.ctx.MarshalBinary(); err != nil {
		return nil, err
	}
	data = append(data, buff...)
	// missing type and size
	return data, nil
}

type psiResponse struct {
	serverSetNum int
	ctxs         []*bfv.Ciphertext
}

func (resp psiResponse) MarshalBinary() (data []byte, err error) {
	var buff []byte
	for _, ctx := range resp.ctxs {
		if buff, err = ctx.MarshalBinary(); err != nil {
			return nil, err
		}
		data = append(data, buff...)
	}
	return data, nil
}
