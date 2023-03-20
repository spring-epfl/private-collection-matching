package psm

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/ldsec/lattigo/v2/bfv"
)

func (cl *client) describeCiphertxt(ctx *bfv.Ciphertext, shorten bool) string {
	ptx := cl.decryptor.DecryptNew(ctx)
	data := cl.encoder.DecodeUintNew(ptx)
	if shorten {
		return fmt.Sprint(data[:8], "...", data[len(data)/2-8:len(data)/2])
	} else {
		return fmt.Sprint(data)
	}
}

func Mod(a, b int) uint64 {
	a = a % b
	if a < 0 {
		a += b
	}
	return uint64(a)
}
func Abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// Intersection(a, b) returns the intersection of the arrays `a` and `b`
func Intersection(a, b []uint64) []uint64 {
	m := make(map[uint64]bool)
	var inter []uint64
	for _, item := range b {
		m[item] = true
	}
	for _, item := range a {
		if _, ok := m[item]; ok {
			inter = append(inter, item)
		}
	}
	return inter
}

// RandomSet(size, max) creates an array of length `size` filled with random elements at most `max`.
// Bad performance if (max - size) is small
// No zeroes
func RandomSet(size int, max int) ([]uint64, error) {
	if size > max {
		s := fmt.Sprintf("Cannot creat a set of %v element with max limit %v", size, max)
		return nil, errors.New(s)
	}

	mapTmp := map[uint64]bool{}
	set := make([]uint64, 0, size)

	for len(set) < size {
		r := rand.Uint64()%uint64(max-1) + 1
		_, in := mapTmp[r]
		if !in {
			mapTmp[r] = true
			set = append(set, r)
		}
	}
	return set, nil
}

func RandomDataSet(setNum, minSetSize, maxSetSize, maxValue int) ([][]uint64, error) {
	sets := make([][]uint64, setNum)
	var err error = nil

	for i := 0; i < setNum; i++ {
		size := minSetSize + (rand.Int() % (maxSetSize - minSetSize + 1))
		sets[i], err = RandomSet(size, maxValue)
		if err != nil {
			return nil, err
		}
	}
	return sets, nil
}

func EncodeSetAsBitVector(set []uint64, dest []uint64) error {
	for i := 0; i < len(dest); i++ {
		dest[i] = 0
	}

	for _, v := range set {
		if int(v) >= len(dest) {
			return errors.New("small domain query inputs must fit in the domain")
		}
		dest[int(v)] = 1
	}
	return nil
}

func EncodeSetsAsBitVector(sets [][]uint64, bitLen int, dest []uint64) error {
	for i := 0; i < len(dest); i++ {
		dest[i] = 0
	}

	for i, set := range sets {
		for _, v := range set {
			if int(v) >= bitLen {
				return errors.New("small domain query inputs must fit in the domain")
			}
			dest[i*bitLen+int(v)] = 1
		}
	}

	return nil
}

func PlainTversky(set1 []uint64, set2 []uint64) int {
	var a, b, c int = 9, 4, 4

	I := Intersection(set1, set2)
	return a*len(I) - b*len(set1) - c*len(set2)
}

func PlainTverskyArray(client []uint64, servers [][]uint64) []int {
	ans := make([]int, len(servers))
	for i := 0; i < len(servers); i++ {
		ans[i] = PlainTversky(client, servers[i])
	}
	return ans
}

func RotatePlainVec(elems []uint64, rot int) []uint64 {
	out := make([]uint64, len(elems))
	rown := len(elems) / 2

	for i := 0; i < rown; i++ {
		out[i] = elems[Mod(i+rot, rown)]
		out[rown+i] = elems[rown+int(Mod(i+rot, rown))]
	}
	return out
}

func ReversePlainVec(elems []uint64) []uint64 {
	out := make([]uint64, len(elems))
	rown := len(elems) / 2

	for i := 0; i < rown; i++ {
		out[i] = elems[Mod(-i, rown)]
		out[rown+i] = elems[rown+int(Mod(-i, rown))]
	}
	return out
}

func GenRandomVector(size, max uint64, allowZero bool) []uint64 {
	data := make([]uint64, size)
	for i := 0; i < int(size); i++ {
		// Warning: USE CRYPTO SECURE RANDOMNESS
		if allowZero {
			data[i] = rand.Uint64() % max
		} else {
			data[i] = (rand.Uint64() % (max - 1)) + 1
		}
	}
	return data
}

func GenRandomPtx(params *bfv.Parameters, allowZero bool) *bfv.PlaintextMul {
	encoder := bfv.NewEncoder(params)

	randData := GenRandomVector(params.N(), params.T(), allowZero)
	randPtx := bfv.NewPlaintextMul(params)
	encoder.EncodeUintMul(randData, randPtx)

	return randPtx
}

// Randomizes a ciphertext in a multiplicative way.
// Zero elements remain the same while any non-zero element gets randomized.
func RandomizeMltCtx(pp *PSIParams, evaluator bfv.Evaluator, ctx *bfv.Ciphertext) *bfv.Ciphertext {
	rPtx := GenRandomPtx(pp.params, false)
	out := evaluator.MulNew(ctx, rPtx)
	evaluator.Relinearize(out, out)
	return out
}

func rearrangeDecryptedBatchedCipher(pp *PSIParams, elems []uint64, batchSize int) []uint64 {
	out := make([]uint64, len(elems))
	N := int(pp.params.N())

	setsPerCtx := N / batchSize

	now := 0
	for i := 0; i < batchSize; i++ {
		for j := 0; j < setsPerCtx; j++ {
			pos := j*batchSize - i
			if (j%(setsPerCtx/2)) == 0 && i > 0 {
				pos += N / 2
			}
			out[now] = elems[pos]
			now++
		}
	}
	return out
}

func rearrangeFPSIResp(resp []uint64, pp *PSIParams) []uint64 {
	N := int(pp.params.N())
	rowN := N / 2

	batchSize := rowN / pp.ClRepNum
	setsPerRow := rowN / batchSize

	out := make([]uint64, N)

	for i := 0; i < batchSize; i++ {
		for j := 0; j < setsPerRow; j++ {
			// fmt.Printf("Raearrange out[%v] <- resp[%v] (%v).\n", i*setsPerRow+j, (rowN+j*batchSize-i)%rowN, resp[(rowN+j*batchSize-i)%rowN])
			// fmt.Printf("Raearrange out[%v] <- resp[%v] (%v).\n", i*setsPerRow+j+rowN, (rowN+j*batchSize-i)%rowN+rowN, resp[(rowN+j*batchSize-i)%rowN+rowN])
			out[i*setsPerRow+j] = resp[(rowN+j*batchSize-i)%rowN]
			out[i*setsPerRow+j+rowN] = resp[(rowN+j*batchSize-i)%rowN+rowN]
		}
	}
	return out
}

func createFPSImask(l int, pp *PSIParams) []uint64 {
	N := int(pp.params.N())
	rowN := N / 2
	batchSize := rowN / pp.ClRepNum
	setsPerRow := rowN / batchSize

	out := make([]uint64, N)
	for i := 0; i < batchSize; i++ {
		for j := 0; j < setsPerRow; j++ {
			if (i*setsPerRow + j) >= l {
				out[(rowN+j*batchSize-i)%rowN] = 1
			}
			if (i*setsPerRow + j + rowN) >= l {
				out[(rowN+j*batchSize-i)%rowN+rowN] = 1
			}
		}
	}
	return out
}

func convertToInt(pp *PSIParams, elems []uint64) []int {
	out := make([]int, len(elems))
	T := int(pp.params.T())

	for i := 0; i < len(elems); i++ {
		out[i] = int(elems[i])
		if out[i] > T/2 {
			out[i] = (out[i] - T) % T
		}
	}
	return out
}

func InterpolateFromRoots(pp *PSIParams, roots []uint64) []uint64 {
	T := pp.params.T()
	Ti := int(T)

	a := make([]int, len(roots)+1)
	a[0] = 1

	for k := 0; k < len(roots); k++ {
		for i := k + 1; i >= 1; i-- {
			a[i] = (a[i-1] - a[i]*int(roots[k])) % Ti
		}
		a[0] = (a[0] * -int(roots[k])) % Ti
	}

	ua := make([]uint64, len(roots)+1)

	for i := 0; i < len(a); i++ {
		if a[i] < 0 {
			a[i] += Ti
		}
		ua[i] = uint64(a[i])
	}
	return ua
}

func MultPlainPloyWithRand(pp *PSIParams, poly []uint64) []uint64 {
	T := pp.params.T()

	out := make([]uint64, len(poly))
	r := rand.Uint64()%uint64(T-1) + 1

	for i := 0; i < len(poly); i++ {
		out[i] = (poly[i] * r) % T
	}

	return out
}

// How many batches (with m slots) is needed to fit x elements
func FitLen(x, m int) int {
	return (x + m - 1) / m
}

func ReadCompoundsFromFile(path string, chemNum int) [][]uint64 {
	sets := make([][]uint64, chemNum)

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)

	for i := 0; i < chemNum; i++ {
		scanner.Scan()
		fp := scanner.Text()
		sets[i] = make([]uint64, 0, 100)
		for j := 0; j < len(fp); j++ {
			if fp[j] == '1' {
				sets[i] = append(sets[i], uint64(j))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return sets
}

func isUintZero(a []uint64) []uint64 {
	out := make([]uint64, len(a))
	for i, v := range a {
		if v == 0 {
			out[i] = 1
		} else {
			out[i] = 0
		}
	}
	return out
}
