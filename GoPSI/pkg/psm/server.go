package psm

import (
	"errors"
	"math/rand"

	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/schollz/progressbar/v3"
)

type server struct {
	pp      *PSIParams
	encoder bfv.Encoder
	N       int

	sets     [][]uint64
	raw_sets [][]uint64
	// set_ptx *bfv.Plaintext

	// Does not support concurrency at the moment
	encryptor bfv.Encryptor
	evaluator bfv.Evaluator
}

func NewServer(pp *PSIParams, sets [][]uint64) (*server, error) {
	params := pp.params
	encoder := bfv.NewEncoder(params)

	N := int(params.N())

	return &server{
		pp:       pp,
		encoder:  encoder,
		N:        N,
		raw_sets: sets,
		sets:     nil,
		// set_ptx: nil,
	}, nil
}

func (sv *server) ShuffleSets() {
	sv.sets = make([][]uint64, len(sv.raw_sets))
	perm := rand.Perm(len(sv.raw_sets))
	for i := range perm {
		sv.sets[i] = sv.raw_sets[perm[i]]
	}
}

func (sv *server) prepareForQuery(key *clientKey) {
	sv.encryptor = bfv.NewEncryptorFromPk(sv.pp.params, key.pk)
	sv.evaluator = bfv.NewEvaluator(sv.pp.params, *key.evk)
}

func (sv *server) Respond(query *psiQuery, key *clientKey) (*psiResponse, error) {
	sv.prepareForQuery(key)

	var resp psiResponse
	qt := query.queryType
	var ctxs []*bfv.Ciphertext

	if qt.Aggregation == AGGREGATION_CA_MS {
		sv.ShuffleSets()
	} else {
		sv.sets = sv.raw_sets
	}

	// Single-set layer
	if qt.IsSmallDomain {
		Logger.Info().Msgf("server: running small domain psi")
		if qt.Psi == PSI_CA {
			Logger.Info().Msgf("server: computing psi-ca")
			var err error
			ctxs, err = sv.computePSI_CA_SD(query)
			if err != nil {
				return nil, err
			}
			if qt.Matching == MATCHING_NONE {
				// If last layer: batch cardinalities into the minimal number of ctxs
				ctxs = BatchSIMDctxs(sv.pp, sv.evaluator, ctxs, sv.pp.SdBitVecLen)
				Logger.Debug().Msgf("Number of batched cardinality ciphertexts: %v", len(ctxs))
			}
		}
	} else if !qt.IsSmallDomain {
		Logger.Info().Msgf("server: running large domain psi")
		var err error
		ctxs, err = sv.interpolationPSI(query)
		if err != nil {
			return nil, err
		}

		if qt.Psi == PSI_CA {
			return nil, errors.New("currently epsi-ca is not supported in the large domain")
		}
	}

	// PSM layer
	if qt.Matching == MATCHING_FPSM {
		Logger.Info().Msgf("server: running f-psm")

		sv.evalFPSM(ctxs)
		ctxs = sv.batchPSMresps(ctxs)
	} else if qt.Matching == MATCHING_TVERSKY || qt.Matching == MATCHING_TVERSKY_PLAIN {
		Logger.Info().Msgf("server: running tversky.")
		// compute plain tversky score
		tvCtx := sv.computeTversky(query, ctxs)
		Logger.Debug().Msgf("Number of Tv ciphertexts: %v", len(ctxs))
		// batch scores into the minimal number of ctxs
		ctxs = BatchSIMDctxs(sv.pp, sv.evaluator, tvCtx, sv.pp.SdBitVecLen)
		Logger.Debug().Msgf("Number of batched Tv ciphertexts: %v", len(ctxs))

		// Convert plain score into binary matching result
		if qt.Matching != MATCHING_TVERSKY_PLAIN {
			Logger.Info().Msgf("server: convert tversky scores to binary matching.")
			sv.convertTverskyScoreToBinary(ctxs, MAX_TVERSKY_SCORE)
		}
	}

	// Many-set layer
	if qt.Aggregation == AGGREGATION_X_MS {
		Logger.Info().Msgf("server: running x-ms aggregation")
		if qt.Matching == MATCHING_FPSM {
			sv.aggregateFPSM(ctxs)
		} else if qt.Matching == MATCHING_TVERSKY {
			ctxs = sv.aggregateTversky(ctxs)
		}
	} else if qt.Aggregation == AGGREGATION_CA_MS {
		Logger.Info().Msgf("server: running ca-ms aggregation")
	}

	// add malicious check
	if qt.IsSmallDomain {
		malCheck := SDMaliciousCheck(sv.pp, sv.evaluator, query.ctx)
		for i := 0; i < len(ctxs); i++ {
			check := RandomizeMltCtx(sv.pp, sv.evaluator, malCheck)
			ctxs[i] = sv.evaluator.AddNew(ctxs[i], check)
		}
	} else {
		malCheck := PolynomialMaliciousCheck(sv.pp, sv.evaluator, query.ctx)
		for i := range ctxs {
			sv.evaluator.Add(ctxs[i], malCheck, ctxs[i])
		}
	}

	resp = psiResponse{
		serverSetNum: len(sv.sets),
		ctxs:         ctxs,
	}
	return &resp, nil
}

//////////////////////////////////
//     Single-set protocols     //
//////////////////////////////////

func (sv *server) computePSI_CA_SD(query *psiQuery) ([]*bfv.Ciphertext, error) {
	// totalCipherNum: Number of input ciphertexts
	totalCipherNum := (len(sv.sets) + sv.pp.sdSetsPerCtx - 1) / sv.pp.sdSetsPerCtx
	caCtx := make([]*bfv.Ciphertext, 0, totalCipherNum)

	// shard: number of (repacked) output ciphertexts
	for shard := 0; shard*sv.N < len(sv.sets); shard++ {

		// select shard's server sets.
		end := (shard + 1) * int(sv.pp.params.N())
		if end > len(sv.sets) {
			end = len(sv.sets)
		}

		sets := sv.sets[shard*sv.N : end]
		cipherNum := (len(sets) + sv.pp.sdSetsPerCtx - 1) / sv.pp.sdSetsPerCtx
		var bar *progressbar.ProgressBar
		if ENABLE_PROGRESS_BAR {
			bar = progressbar.Default(int64(len(sets)), "Intersection progress")
		}

		for k := 0; k < cipherNum; k++ {
			bitVec := make([]uint64, sv.pp.params.N())
			next := (k + 1) * sv.pp.sdSetsPerCtx
			if next > len(sets) {
				next = len(sets)
			}
			err := EncodeSetsAsBitVector(sets[k*sv.pp.sdSetsPerCtx:next], sv.pp.SdBitVecLen, bitVec)
			if err != nil {
				return nil, err
			}

			selectPtx := bfv.NewPlaintextMul(sv.pp.params)
			sv.encoder.EncodeUintMul(bitVec, selectPtx)
			selCtx := sv.evaluator.MulNew(query.ctx, selectPtx)
			SumSIMD(sv.evaluator, selCtx, sv.pp.SdBitVecLen)

			// IMPORTANT not secure for simple cardinality -> improves noise for tversky
			// caCtx[k] = FilterSIMD(sv.evaluator, selCtx, sv.pp.sdBitVecLen)
			caCtx = append(caCtx, selCtx)

			if ENABLE_PROGRESS_BAR {
				bar.Set(next)
			}
		}
	}

	return caCtx, nil
}

func (sv *server) interpolationPSI(query *psiQuery) ([]*bfv.Ciphertext, error) {
	ctxs := make([]*bfv.Ciphertext, FitLen(len(sv.sets), sv.pp.ClRepNum))
	rowN := int(sv.pp.params.N()) / 2

	ptx := bfv.NewPlaintextMul(sv.pp.params)

	for cn := 0; cn < len(ctxs); cn++ {
		expandedSet := make([]uint64, sv.pp.params.N())

		for rep := 0; rep < sv.pp.ClRepNum; rep++ {
			n := cn*sv.pp.ClRepNum + rep
			if n >= len(sv.sets) {
				continue
			}
			if len(sv.sets[n]) > sv.pp.ClientPolyExpansion-1 {
				return nil, errors.New("too many elements in one of server sets")
			}

			// Note:
			// Interpolation works as: a[0]*1 + a[1]*x + a[2]*x^2 ...
			// Client packs input as: c, c^2, c^3, ...
			// The starting difference 1 vs c acts as adding (x == 0) to roots
			a := InterpolateFromRoots(sv.pp, sv.sets[n])
			// randomize a for each use

			for k := 0; k < sv.pp.MaxClientElemPerCtx/2; k++ {
				base := (sv.pp.MaxClientElemPerCtx/2*rep + k) * sv.pp.ClientPolyExpansion
				ar1, ar2 := MultPlainPloyWithRand(sv.pp, a), MultPlainPloyWithRand(sv.pp, a)

				for i := 0; i < len(a); i++ {
					expandedSet[base+i] = ar1[i]
					expandedSet[rowN+base+i] = ar2[i]
				}
			}
		}

		sv.encoder.EncodeUintMul(expandedSet, ptx)
		ctxs[cn] = sv.evaluator.MulNew(query.ctx, ptx)
		SumSIMD(sv.evaluator, ctxs[cn], sv.pp.ClientPolyExpansion)
	}
	return ctxs, nil
}

//////////////////////////////////
//        PSM protocols         //
//////////////////////////////////

func (sv *server) evalFPSM(psi []*bfv.Ciphertext) {
	params := sv.pp.params
	rowN := int(params.N()) / 2

	raw := make([]uint64, params.N())
	ptx := bfv.NewPlaintextMul(params)
	batchSize := rowN / sv.pp.ClRepNum

	for k := 0; k < len(psi); k++ {
		if k%10 == 0 {
			Logger.Debug().Msgf("Running FPSM %v.", k)
		}
		psi[k] = SIMDOperation(sv.evaluator, psi[k],
			sv.pp.ClientPolyExpansion,
			rowN/sv.pp.ClRepNum,
			true, false)

		// Randomizes non-zero c[0] and zero out everything else
		// Assumes one set per ctx
		for i := 0; i < sv.pp.ClRepNum; i++ {
			raw[i*batchSize] = (rand.Uint64() % (params.T() - 1)) + 1
		}
		sv.encoder.EncodeUintMul(raw, ptx)

		sv.evaluator.Mul(psi[k], ptx, psi[k])
		sv.evaluator.Relinearize(psi[k], psi[k])
	}
}

func (sv *server) batchPSMresps(psm []*bfv.Ciphertext) (ctxs []*bfv.Ciphertext) {
	batchSize := sv.N / 2 / sv.pp.ClRepNum

	ctxs = make([]*bfv.Ciphertext, FitLen(len(psm), 2*batchSize))

	for k := 0; k < len(ctxs); k++ {
		end := 2 * batchSize * (k + 1)
		if end > len(psm) {
			end = len(psm)
		}
		now := psm[k*2*batchSize : end]
		if len(now) > batchSize {
			up := LinearBatch(sv.evaluator, now[:batchSize])
			down := LinearBatch(sv.evaluator, now[batchSize:])
			down = sv.evaluator.RotateRowsNew(down)
			ctxs[k] = sv.evaluator.AddNew(up, down)
		} else {
			ctxs[k] = LinearBatch(sv.evaluator, now)
		}
	}

	// Set the value of empty sets to 1.
	// Note that empty sets are guaranteed to have PSM output equal to zero.
	mask := createFPSImask(len(sv.sets)%sv.N, sv.pp)
	maskPtx := bfv.NewPlaintext(sv.pp.params)
	sv.encoder.EncodeUint(mask, maskPtx)
	sv.evaluator.Add(ctxs[len(ctxs)-1], maskPtx, ctxs[len(ctxs)-1])

	return ctxs
}

func (sv *server) computeTversky(query *psiQuery, intersectionCaCtx []*bfv.Ciphertext) []*bfv.Ciphertext {
	// Tversky params for \alpha = \beta = 1, t = 80%
	var a, b, c uint64 = 9, 4, 4

	tvCtx := make([]*bfv.Ciphertext, len(intersectionCaCtx))

	// compute the cardinality of the client's set |X|. (Different from intersection cardinality)
	clientCaCtx := query.ctx.CopyNew().Ciphertext()
	SumSIMD(sv.evaluator, clientCaCtx, sv.pp.SdBitVecLen)
	sv.evaluator.MulScalar(clientCaCtx, b, clientCaCtx)

	for k := 0; k < len(intersectionCaCtx); k++ {
		sv.evaluator.MulScalar(intersectionCaCtx[k], a, intersectionCaCtx[k])

		// set server sets' cardinality |S_i|
		serverCaRaw := make([]uint64, sv.pp.params.N())
		for i := 0; i < sv.pp.sdSetsPerCtx; i++ {
			if k*sv.pp.sdSetsPerCtx+i < len(sv.sets) {
				serverCaRaw[i*sv.pp.SdBitVecLen] = c * uint64(len(sv.sets[k*sv.pp.sdSetsPerCtx+i]))
			}
		}
		serverCaPtx := bfv.NewPlaintext(sv.pp.params)
		sv.encoder.EncodeUint(serverCaRaw, serverCaPtx)

		tmp := sv.evaluator.AddNew(serverCaPtx, clientCaCtx)
		tvCtx[k] = sv.evaluator.SubNew(intersectionCaCtx[k], tmp)
	}
	return tvCtx
}

func (sv *server) convertTverskyScoreToBinary(tvCtx []*bfv.Ciphertext, scoreLim int) {
	for i := 0; i < len(tvCtx); i++ {
		// IMPORTANT range support varies with noise bidget
		tvCtx[i] = IsInRange(sv.pp, sv.evaluator, tvCtx[i], scoreLim)

		// randomizing Tversky out to ensure privacy
		rPtx := GenRandomPtx(sv.pp.params, false)
		sv.evaluator.Mul(tvCtx[i], rPtx, tvCtx[i])
	}
}

//////////////////////////////////
//     Many-set aggregation     //
//////////////////////////////////

func (sv *server) aggregateFPSM(ctxs []*bfv.Ciphertext) error {
	// ONLY SUPPORTS 1 CTX
	if len(ctxs)*sv.pp.ClRepNum > sv.N {
		return errors.New("too many server sets")
	}

	batchSize := sv.N / 2 / sv.pp.ClRepNum
	rotateToPositive := len(ctxs) - 1
	ctxs[0] = ExtendedRotate(sv.pp, sv.evaluator, ctxs[0], -rotateToPositive)
	ctxs[0] = SIMDOperation(sv.evaluator, ctxs[0], 1, len(ctxs), false, true)
	ctxs[0] = SIMDOperation(sv.evaluator, ctxs[0], batchSize, sv.N/2, false, true)
	return nil
}

func (sv *server) aggregateTversky(ctxs []*bfv.Ciphertext) []*bfv.Ciphertext {
	maxMultDept := 64
	if len(ctxs) == 1 {
		// internal aggregation when only one response ctx exists
		ctxs[0] = SIMDOperation(sv.evaluator, ctxs[0], 256, 64*256, false, true)
	} else {
		// aggregate ctxs together
		tmp := make([]*bfv.Ciphertext, FitLen(len(ctxs), maxMultDept))
		for k := 0; k*maxMultDept < len(ctxs); k++ {
			start, end := k*maxMultDept, (k+1)*maxMultDept
			if end > len(ctxs) {
				end = len(ctxs)
			}
			Logger.Debug().Msgf("Aggregate ciphers %v - %v.", start, end)
			tmp[k] = ArrayOperation(sv.evaluator, ctxs[start:end], true)
		}
		ctxs = tmp
	}
	return ctxs
}

// //////////////////////////
//
//	Malicious check        //
//
// //////////////////////////
func PolynomialMaliciousCheck(pp *PSIParams, evaluator bfv.Evaluator, poly *bfv.Ciphertext) *bfv.Ciphertext {
	// Computes P.rShifted == [P.rRaw.(cc-cn)].rot(-1)

	encoder := bfv.NewEncoder(pp.params)

	rVec := GenRandomVector(pp.params.N(), pp.params.T(), false)
	rRawPtx := bfv.NewPlaintext(pp.params)
	encoder.EncodeUint(rVec, rRawPtx)
	for i := pp.params.N() - 1; i > 0; i-- {
		rVec[i] = rVec[i-1]
	}
	for i := 0; i < int(pp.params.N()); i += pp.ClientPolyExpansion {
		rVec[i] = 0
	}
	rShiftedPtx := bfv.NewPlaintext(pp.params)
	encoder.EncodeUint(rVec, rShiftedPtx)

	c0 := FilterSIMD(pp, evaluator, poly, pp.ClientPolyExpansion)     //c in c[0]
	cn := ExtendedRotate(pp, evaluator, c0, 1-pp.ClientPolyExpansion) //c in c[-1]
	cc := SIMDOperation(evaluator, c0, 1, pp.ClientPolyExpansion, false, false)
	cc = ExtendedRotate(pp, evaluator, cc, 1-pp.ClientPolyExpansion) // c in c[:]
	ccMinCn := evaluator.SubNew(cc, cn)                              // c in c[:-1]

	left := evaluator.MulNew(poly, rShiftedPtx)
	right := evaluator.MulNew(poly, rRawPtx)
	right = evaluator.MulNew(right, ccMinCn)
	evaluator.Relinearize(right, right)
	evaluator.RotateColumns(right, -1, right)
	powerCheck := evaluator.SubNew(left, right)
	malCheck := powerCheck

	if pp.ClRepNum > 1 {
		polyRepRot := evaluator.RotateColumnsNew(poly, int(pp.params.N())/2/pp.ClRepNum)
		duplicateCheck := evaluator.SubNew(poly, polyRepRot)
		dupR := GenRandomPtx(pp.params, false)
		duplicateCheck = evaluator.MulNew(duplicateCheck, dupR)
		evaluator.Relinearize(duplicateCheck, duplicateCheck)

		malCheck = evaluator.AddNew(powerCheck, duplicateCheck)
	}

	malCheck = SIMDOperation(evaluator, malCheck, 1, int(pp.params.N()/2), true, false)
	finalR := GenRandomPtx(pp.params, false)
	evaluator.Mul(malCheck, finalR, malCheck)
	evaluator.Relinearize(malCheck, malCheck)

	return malCheck
}

// MalCheck MUST get re-randomized before use
func SDMaliciousCheck(pp *PSIParams, evaluator bfv.Evaluator, q *bfv.Ciphertext) *bfv.Ciphertext {
	qMinOne := evaluator.SubNew(q, pp.rangePtxs[1])
	sdCheck := evaluator.MulNew(q, qMinOne)
	evaluator.Relinearize(sdCheck, sdCheck)
	sdCheck = RandomizeMltCtx(pp, evaluator, sdCheck)

	qRepRot := evaluator.RotateColumnsNew(q, pp.SdBitVecLen)
	duplicateCheck := evaluator.SubNew(q, qRepRot)
	duplicateCheck = RandomizeMltCtx(pp, evaluator, duplicateCheck)

	malCheck := evaluator.AddNew(sdCheck, duplicateCheck)
	malCheck = SIMDOperation(evaluator, malCheck, 1, int(pp.params.N()/2), true, false)
	return malCheck
}
