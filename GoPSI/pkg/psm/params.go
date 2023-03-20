package psm

import (
	"fmt"

	"github.com/ldsec/lattigo/v2/bfv"
)

type PSIParams struct {
	params *bfv.Parameters

	// large domain parameters
	MaxClientElemPerCtx int // must be in form of 2^k
	ClRepNum            int // must be in form of 2^k
	ClientPolyExpansion int // Determines the polynomial expansion degree of the client
	// server set sizes must be smaller than clientExpansion
	// clientPolyExpansion can be computed knowing N, ClRepNum, and maxClientElemPerCtx

	// small domain parameters
	SdBitVecLen  int // must be in form of 2^k
	sdSetsPerCtx int

	onePtx    *bfv.PlaintextMul
	zeroPtx   *bfv.PlaintextMul
	rangePtxs []*bfv.Plaintext
}

func NewPSIParams(params *bfv.Parameters, rangeLim int) *PSIParams {
	pp := &PSIParams{
		params:              params,
		ClRepNum:            1,
		MaxClientElemPerCtx: 16,
		SdBitVecLen:         256,
	}
	encoder := bfv.NewEncoder(params)

	pp.Update()

	pp.onePtx = bfv.NewPlaintextMul(params)
	pp.zeroPtx = bfv.NewPlaintextMul(params)

	vec := make([]uint64, params.N())
	encoder.EncodeUintMul(vec, pp.zeroPtx)
	for i := 0; i < len(vec); i++ {
		vec[i] = 1
	}
	encoder.EncodeUintMul(vec, pp.onePtx)

	pp.rangePtxs = make([]*bfv.Plaintext, rangeLim)
	for k := 0; k < rangeLim; k++ {
		pp.rangePtxs[k] = bfv.NewPlaintext(params)
		for i := 0; i < len(vec); i++ {
			vec[i] = uint64(k)
		}
		encoder.EncodeUint(vec, pp.rangePtxs[k])
	}

	return pp
}

func (pp *PSIParams) Update() {
	pp.ClientPolyExpansion = int(pp.params.N()) / pp.MaxClientElemPerCtx / pp.ClRepNum
	pp.sdSetsPerCtx = int(pp.params.N()) / pp.SdBitVecLen
}

func (pp *PSIParams) Describe() string {
	desc := ""
	desc += fmt.Sprintf("Number of query replicates in the ciphertext: %v\n", pp.ClRepNum)
	desc += fmt.Sprintf("Small domain => SdBitVecLen: %v\n", pp.SdBitVecLen)
	desc += fmt.Sprintf("Small input  => Max client element per ctx: %v, max server size: %v", pp.MaxClientElemPerCtx, pp.ClientPolyExpansion)
	return desc
}

func GetBFVParam(logn int) (params *bfv.Parameters) {
	switch logn {
	case 12:
		params = bfv.DefaultParams[bfv.PN12QP109].WithT(40961)
	case 13:
		params = bfv.DefaultParams[bfv.PN13QP218].WithT(4079617) // 22 bit T -> depth:3
		// params = bfv.DefaultParams[bfv.PN13QP218].WithT(1073692673) // 30 bit T  -> depth:2
	case 14:
		params = bfv.DefaultParams[bfv.PN14QP438].WithT(163841)
	case 15:
		params = bfv.DefaultParams[bfv.PN15QP880].WithT(786433)
	}
	return params
}

// function taken from lattigo samples
func DescribeParams(params *bfv.Parameters) {
	fmt.Println("================== Parameters ==================")
	fmt.Printf("Parameters : N=%d, T=%d, Q = %d bits, sigma = %f \n",
		1<<params.LogN(), params.T(), params.LogQP(), params.Sigma())
	fmt.Println("================================================")
}

func testDepth(params *bfv.Parameters) {
	pp := NewPSIParams(params, 101)

	cl := NewClient(pp)
	evaluator := bfv.NewEvaluator(cl.pp.params, *cl.evk)

	data := []uint64{1, 2, 3, 4}
	ptx := bfv.NewPlaintext(cl.pp.params)
	cl.encoder.EncodeUint(data, ptx)
	ctx := cl.encryptor.EncryptNew(ptx)

	for i := 1; i < 20; i++ {
		ctx = evaluator.MulNew(ctx, ctx)
		evaluator.Relinearize(ctx, ctx)
		fmt.Printf("[Depth %v] -> %v\n", i, cl.describeCiphertxt(ctx, true))
	}
}
