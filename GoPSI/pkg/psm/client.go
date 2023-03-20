package psm

import (
	"errors"

	"github.com/ldsec/lattigo/v2/bfv"
)

// NOTE: Fixing params can allow more preprocess
type client struct {
	pp        *PSIParams
	pk        *bfv.PublicKey
	evk       *bfv.EvaluationKey
	sk        *bfv.SecretKey
	encoder   bfv.Encoder
	encryptor bfv.Encryptor
	decryptor bfv.Decryptor
}

func NewClient(pp *PSIParams) *client {
	cl := new(client)
	cl.pp = pp
	params := pp.params

	// Needs to be in sync with param and operation
	rots := make([]int, 0, 30)
	for k := 1; k < int(params.N()); k *= 2 {
		rots = append(rots, k)
	}
	rots = append(rots, -1)

	keyGen := bfv.NewKeyGenerator(params)
	cl.sk, cl.pk = keyGen.GenKeyPair()
	rlk := keyGen.GenRelinearizationKey(cl.sk, 2)
	// include row swap and all pow(2)
	rtk := keyGen.GenRotationKeysForRotations(rots, true, cl.sk)
	cl.evk = &bfv.EvaluationKey{
		Rlk:  rlk,
		Rtks: rtk,
	}

	cl.encoder = bfv.NewEncoder(params)
	cl.encryptor = bfv.NewEncryptorFromSk(params, cl.sk)
	cl.decryptor = bfv.NewDecryptor(params, cl.sk)

	return cl
}

func (cl *client) GetKey() *clientKey {
	key := clientKey{
		pk:  cl.pk,
		evk: cl.evk,
	}
	return &key
}

func (cl *client) Query(set []uint64, queryType QueryType) (*psiQuery, error) {
	expandedSet := make([]uint64, cl.pp.params.N())

	if queryType.IsSmallDomain {
		// replicates the bit vector till it fills all the slots
		// sdBitVecLen is a power of 2
		Logger.Info().Msgf("Create a small domain query.")
		for i := 0; i < int(cl.pp.params.N())/cl.pp.SdBitVecLen; i++ {
			EncodeSetAsBitVector(set, expandedSet[i*cl.pp.SdBitVecLen:(i+1)*cl.pp.SdBitVecLen])
		}

	} else {
		// Large domain protocols
		Logger.Info().Msgf("Create a large domain query.")
		Logger.Debug().Msgf("Max client size: %v, Max server size %v, Replica per ctx: %v.", cl.pp.MaxClientElemPerCtx, cl.pp.ClientPolyExpansion, cl.pp.ClRepNum)

		if len(set) > cl.pp.MaxClientElemPerCtx {
			return nil, errors.New("too many client elements in query")
		}

		// Polynomial replication
		for rep := 0; rep < cl.pp.ClRepNum; rep++ {
			for k := 0; k < len(set); k++ {
				base := rep * cl.pp.MaxClientElemPerCtx / 2
				base += k % (cl.pp.MaxClientElemPerCtx / 2)
				base *= cl.pp.ClientPolyExpansion
				if k >= (cl.pp.MaxClientElemPerCtx / 2) {
					base += int(cl.pp.params.N()) / 2
				}

				expandedSet[base] = set[k]
				for i := 1; i < cl.pp.ClientPolyExpansion; i++ {
					expandedSet[base+i] = (expandedSet[base+i-1] * set[k]) % cl.pp.params.T()
				}
			}
		}
	}

	ptx := bfv.NewPlaintext(cl.pp.params)
	cl.encoder.EncodeUint(expandedSet, ptx)
	ctx := cl.encryptor.EncryptNew(ptx)

	q := psiQuery{
		clientSetSize: len(set),
		ctx:           ctx,
		queryType:     queryType,
	}
	return &q, nil
}

func (cl *client) EvalResponse(clientSet []uint64, query *psiQuery, resp *psiResponse) []uint64 {
	Logger.Info().Msgf("client: evaluating the response")

	qt := query.queryType

	// Only PSI layer
	if qt.Matching == MATCHING_NONE {
		Logger.Info().Msgf("client: simple psi layer without matching or aggregation.")

		if qt.Psi == PSI_PSI {
			intersections := make([][]uint64, len(resp.ctxs))
			for k := 0; k < len(resp.ctxs); k++ {
				respPtx := cl.decryptor.DecryptNew(resp.ctxs[k])
				respData := cl.encoder.DecodeUintNew(respPtx)
				intersections[k] = make([]uint64, 0, len(clientSet))
				for i, v := range clientSet {
					if respData[i*cl.pp.ClientPolyExpansion] == 0 {
						intersections[k] = append(intersections[k], v)
					}
				}
			}
			// Warning: For API compatibility, we only return the intersection with the first set since the output type is []uint64
			return intersections[0]
		} else if qt.Psi == PSI_CA && qt.IsSmallDomain {
			ans := make([]uint64, 0, resp.serverSetNum)
			for _, ctx := range resp.ctxs {
				respPtx := cl.decryptor.DecryptNew(ctx)
				respData := cl.encoder.DecodeUintNew(respPtx)
				respData = rearrangeDecryptedBatchedCipher(cl.pp, respData, cl.pp.SdBitVecLen)
				ans = append(ans, respData...)
			}
			if len(ans) > resp.serverSetNum {
				ans = ans[:resp.serverSetNum]
			}
			return ans
		}
	}

	ans := make([]uint64, 0, resp.serverSetNum)
	for _, ctx := range resp.ctxs {
		respPtx := cl.decryptor.DecryptNew(ctx)
		respData := cl.encoder.DecodeUintNew(respPtx)

		if qt.Matching == MATCHING_FPSM {
			respData = rearrangeFPSIResp(respData, cl.pp)
		} else if qt.Matching == MATCHING_TVERSKY || qt.Matching == MATCHING_TVERSKY_PLAIN {
			respData = rearrangeDecryptedBatchedCipher(cl.pp, respData, cl.pp.SdBitVecLen)
		}

		ans = append(ans, respData...)
	}
	if len(ans) > resp.serverSetNum {
		ans = ans[:resp.serverSetNum]
	}

	// No aggregation or zero check for plain tversky
	if qt.Matching == MATCHING_TVERSKY_PLAIN {
		Logger.Info().Msgf("client: plain tversky response.")
		return ans
	} else if qt.Matching != MATCHING_NONE {
		ans = isUintZero(ans)
	}

	if qt.Aggregation == AGGREGATION_NONE {
		Logger.Info().Msgf("client: response without aggregation.")
		return ans
	} else {
		Logger.Info().Msgf("client: aggregated response.")
		if qt.Aggregation == AGGREGATION_CA_MS {
			for _, v := range ans[1:] {
				ans[0] += v
			}
		}
		ans = ans[:1]
		return ans
	}

}
