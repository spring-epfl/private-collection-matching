package psm

import (
	"errors"

	"github.com/ldsec/lattigo/v2/bfv"
)

//////////////////////////////////
//  Arithmetic SIMD operations  //
//////////////////////////////////

// Computes sum_{i \in [0, n)} c[i]
// Only supports n \in pow(2) [upscales n to the next power]
func SumSIMD(evaluator bfv.Evaluator, ctx *bfv.Ciphertext, n int) {
	tmp := ctx.CopyNew().Ciphertext()

	for shift := 1; shift < n; shift *= 2 {
		evaluator.RotateColumns(ctx, shift, tmp)
		evaluator.Add(ctx, tmp, ctx)
	}
}

// Computes mult_{i \in [0, n)} c[i]
// Only supports n \in pow(2) [upscales n to the next power]
func ProdSIMD(evaluator bfv.Evaluator, ctx *bfv.Ciphertext, n int) *bfv.Ciphertext {
	tmp := ctx.CopyNew().Ciphertext()

	for shift := 1; shift < n; shift *= 2 {
		evaluator.RotateColumns(ctx, shift, tmp)
		ctx = evaluator.MulNew(ctx, tmp)
		evaluator.Relinearize(ctx, ctx)
	}
	return ctx
}

func SIMDOperation(evaluator bfv.Evaluator, ctx *bfv.Ciphertext, start, finish int, combineRow, isMul bool) *bfv.Ciphertext {
	tmp := ctx.CopyNew().Ciphertext()
	ctx = tmp.CopyNew().Ciphertext()

	for shift := start; shift < finish; shift *= 2 {
		evaluator.RotateColumns(ctx, shift, tmp)
		if isMul {
			ctx = evaluator.MulNew(ctx, tmp)
			evaluator.Relinearize(ctx, ctx)
		} else {
			evaluator.Add(ctx, tmp, ctx)
		}
	}
	if combineRow {
		evaluator.RotateRows(ctx, tmp)
		if isMul {
			ctx = evaluator.MulNew(ctx, tmp)
			evaluator.Relinearize(ctx, ctx)
		} else {
			evaluator.Add(ctx, tmp, ctx)
		}
	}
	return ctx
}

// Takes an array of ciphertext and reduces it to one ctx by performing element-wise operations
// Returns sum/mult ctx[i]
// Depth optimal
// destroys the array
func ArrayOperation(evaluator bfv.Evaluator, ctxs []*bfv.Ciphertext, isMul bool) *bfv.Ciphertext {
	n := len(ctxs)
	for n > 1 {
		for i := 0; (2*i + 1) < n; i++ {
			if isMul {
				ctxs[i] = evaluator.MulNew(ctxs[2*i], ctxs[2*i+1])
				evaluator.Relinearize(ctxs[i], ctxs[i])
			} else {
				ctxs[i] = evaluator.AddNew(ctxs[2*i], ctxs[2*i+1])
			}
		}
		if n%2 == 1 {
			ctxs[n/2] = ctxs[n-1]
		}
		n = (n + 1) / 2
	}

	return ctxs[0]
}

//////////////////////////////////
//    Selection and filtering   //
//////////////////////////////////

// out[i] = ctx1[i] if sel[i] else ctx0[i]
func SelectSIMD(pp *PSIParams, evaluator bfv.Evaluator, ctx0, ctx1 *bfv.Ciphertext, sel []uint64) *bfv.Ciphertext {
	params := pp.params
	encoder := bfv.NewEncoder(params)

	nsel := make([]uint64, params.N())
	for i := 0; i < int(params.N()); i++ {
		if i < len(sel) {
			nsel[i] = 1 - sel[i]
		} else {
			nsel[i] = 1
		}
	}

	sel_ptx := bfv.NewPlaintextMul(params)
	encoder.EncodeUintMul(sel, sel_ptx)
	nsel_ptx := bfv.NewPlaintextMul(params)
	encoder.EncodeUintMul(nsel, nsel_ptx)

	s0 := evaluator.MulNew(ctx0, nsel_ptx)
	s1 := evaluator.MulNew(ctx1, sel_ptx)

	evaluator.Add(s0, s1, s0)
	return s0
}

// Only keeps c[i] where (i % n) = 0 for n = 2^k
func FilterSIMD(pp *PSIParams, evaluator bfv.Evaluator, ctx *bfv.Ciphertext, n int) *bfv.Ciphertext {
	params := pp.params
	filter := make([]uint64, params.N())
	for i := 0; i < int(params.N()); i += n {
		filter[i] = 1
	}

	filter_ptx := bfv.NewPlaintextMul(params)
	encoder := bfv.NewEncoder(params)
	encoder.EncodeUintMul(filter, filter_ptx)

	return evaluator.MulNew(ctx, filter_ptx)
}

//////////////////////////////////
//     Extended operations      //
//////////////////////////////////

// Checks c[i] /in [0, n)
func IsInRange(pp *PSIParams, evaluator bfv.Evaluator, ctx *bfv.Ciphertext, n int) *bfv.Ciphertext {
	ctxs := make([]*bfv.Ciphertext, n)

	for i := 0; i < n; i++ {
		ctxs[i] = evaluator.SubNew(ctx, pp.rangePtxs[i])
	}

	return ArrayOperation(evaluator, ctxs, true)
}

// Enables rotation with an arbitrary i when only rotation keys in the form of 2^k are available
func ExtendedRotate(pp *PSIParams, evaluator bfv.Evaluator, ctx *bfv.Ciphertext, rot int) *bfv.Ciphertext {
	ans := ctx.CopyNew().Ciphertext()

	if rot < 0 {
		rot += int(pp.params.N() / 2)
	}

	for k := 1; rot > 0; k *= 2 {
		if rot%2 == 1 {
			evaluator.RotateColumns(ans, k, ans)
		}
		rot /= 2
	}

	return ans
}

// Perform element-wise power ^n
func Pow(evaluator bfv.Evaluator, x *bfv.Ciphertext, n int) (*bfv.Ciphertext, error) {
	if n == 0 {
		return nil, errors.New("cannot perform power(x, 0)")
	}

	if n == 1 {
		return x, nil
	}

	xsquare := evaluator.MulNew(x, x)
	evaluator.Relinearize(xsquare, xsquare)
	if n%2 == 0 {
		var pow, err = Pow(evaluator, xsquare, n/2)
		return pow, err
	} else {
		var pow, err = Pow(evaluator, xsquare, (n-1)/2)
		if err != nil {
			return nil, err
		}
		res := evaluator.MulNew(x, pow)
		evaluator.Relinearize(res, res)
		return res, nil
	}
}

func IsZero(evaluator bfv.Evaluator, x *bfv.Ciphertext) (*bfv.Ciphertext, error) {
	return nil, errors.New("IsZero requires depth incompatible with batching")
}

//////////////////////////////////
//           Batching           //
//////////////////////////////////

// Combine ctxs where only c[x%fanIn == 0] has interesting data
// fanIn must b in form of 2^x
func BatchSIMDctxs(pp *PSIParams, evaluator bfv.Evaluator, ctxs []*bfv.Ciphertext, fanIn int) []*bfv.Ciphertext {
	out := make([]*bfv.Ciphertext, 0, len(ctxs)/fanIn+1)

	for processed := 0; processed < len(ctxs); {
		next := fanIn
		if processed+next > len(ctxs) {
			next = len(ctxs) - processed
		}
		for i := 0; i < next; i++ {
			ctxs[processed+i] = FilterSIMD(pp, evaluator, ctxs[processed+i], fanIn)
			ctxs[processed+i] = ExtendedRotate(pp, evaluator, ctxs[processed+i], int(i))
		}
		combined := ArrayOperation(evaluator, ctxs[processed:processed+next], false)
		out = append(out, combined)
		processed += next
	}

	return out
}

// return \sum ctx[i].rot(i)
// Destroys the input array
func LinearBatch(evaluator bfv.Evaluator, ctxs []*bfv.Ciphertext) *bfv.Ciphertext {
	n := len(ctxs)
	rot := 1
	for n > 1 {
		for i := 0; (2*i + 1) < n; i++ {
			// fmt.Printf("rotate c[%v] by %v. Combine c[%v]-c[%v] as c[%v]\n", 2*i+1, rot, 2*i, 2*i+1, i)
			evaluator.RotateColumns(ctxs[2*i+1], rot, ctxs[2*i+1])
			ctxs[i] = evaluator.AddNew(ctxs[2*i], ctxs[2*i+1])
		}
		if n%2 == 1 {
			ctxs[n/2] = ctxs[n-1]
		}
		n = (n + 1) / 2
		rot *= 2
	}

	return ctxs[0]
}
