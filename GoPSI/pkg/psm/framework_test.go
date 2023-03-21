package psm

import (
	"reflect"
	"testing"
)

const PARAM_SIZE = 15

func runHomoPsi(paramSize int, clientSet []uint64, serverSets [][]uint64, qt QueryType, repNum int) (*PSIParams, []uint64) {
	bfvParams := GetBFVParam(paramSize)
	pp := NewPSIParams(bfvParams, 128)
	pp.ClRepNum = repNum
	pp.Update()

	cl := NewClient(pp)
	sv, err := NewServer(pp, serverSets)
	if err != nil {
		panic(err)
	}

	clKey := cl.GetKey()

	query, err := cl.Query(clientSet, qt)
	if err != nil {
		panic(err)
	}
	resp, err := sv.Respond(query, clKey)
	if err != nil {
		panic(err)
	}

	ans := cl.EvalResponse(clientSet, query, resp)

	return pp, ans
}

func checkPlainTversky(paramSize int, clientSet []uint64, serverSets [][]uint64) bool {
	Logger.Debug().Msgf("running check plain tversky")

	qt, err := NewQueryType(true, PSI_CA, MATCHING_TVERSKY_PLAIN, AGGREGATION_NAIVE)
	if err != nil {
		panic(err)
	}

	pp, ans := runHomoPsi(paramSize, clientSet, serverSets, *qt, 1)

	tversky := PlainTverskyArray(clientSet, serverSets)

	T := int(pp.params.T())

	for i, v := range tversky {
		// handle negative numbers
		check := uint64((v + T) % T)
		if check != ans[i] {
			Logger.Warn().Msgf("Set %v is incorrect:", i)
			Logger.Warn().Msgf("  | Client: %v, server: %v", clientSet, serverSets[i])
			Logger.Warn().Msgf("  | |I|: %v, tversky: %v, ans: %v, check: %v, diff: %v",
				len(Intersection(clientSet, serverSets[i])), tversky[i], ans[i], check, int(check-ans[i]))

			return false
		}
	}
	return true
}

func checkTversky(paramSize int, clientSet []uint64, serverSets [][]uint64) bool {
	Logger.Debug().Msgf("running check psm tversky")

	qt, err := NewQueryType(true, PSI_CA, MATCHING_TVERSKY, AGGREGATION_NAIVE)
	if err != nil {
		panic(err)
	}

	_, ans := runHomoPsi(paramSize, clientSet, serverSets, *qt, 1)

	tversky := PlainTverskyArray(clientSet, serverSets)

	for i := range serverSets {
		match := uint64(0)
		if tversky[i] >= 0 {
			match = 1
			Logger.Debug().Msgf("Set %v is a match.", i)
		}

		if match != ans[i] {
			Logger.Warn().Msgf("Set %v is incorrect:", i)
			Logger.Warn().Msgf("  | Client: %v, server: %v", clientSet, serverSets[i])
			Logger.Warn().Msgf("  | |I|: %v, tversky: %v, ans: %v, match %v",
				len(Intersection(clientSet, serverSets[i])), tversky[i], ans[i], match)
			return false
		}
	}
	return true
}

func checkAggregatedTversky(paramSize int, clientSet []uint64, serverSets [][]uint64, agg AggregationType) bool {
	Logger.Debug().Msgf("running check aggregated tversky")

	qt, err := NewQueryType(true, PSI_CA, MATCHING_TVERSKY, agg)
	if err != nil {
		panic(err)
	}

	_, ans := runHomoPsi(paramSize, clientSet, serverSets, *qt, 1)
	tversky := PlainTverskyArray(clientSet, serverSets)
	matches_truth := uint64(0)

	for i := range serverSets {
		if tversky[i] >= 0 {
			matches_truth += 1
		}
	}
	if agg == AGGREGATION_X_MS {
		if matches_truth > 0 {
			matches_truth = 1
		}
	}

	return (ans[0] == matches_truth)
}

func TestTverskySmall(t *testing.T) {
	Logger.Info().Msgf("\n    Test: running TestTverskySmall")
	clientSet := []uint64{1, 2, 3, 4, 5, 6}
	serverSets := [][]uint64{
		{1, 2, 3, 4, 5, 6},
		{1, 2, 3, 4, 5},
		{1, 2, 3, 4, 5, 6, 7},
		{1, 2, 3, 4, 5, 6, 7, 8, 9},
		{10, 20, 30, 40, 50, 60},
	}

	qt, err := NewQueryType(true, PSI_CA, MATCHING_TVERSKY_PLAIN, AGGREGATION_NAIVE)
	if err != nil {
		panic(err)
	}

	_, ans := runHomoPsi(14, clientSet, serverSets, *qt, 1)

	correct := []uint64{6, 1, 2, 163835, 163793}

	if reflect.DeepEqual(ans, correct) == false {
		Logger.Warn().Msgf("ans %v - correct %v", ans, correct)
		t.Error("Mismatch")
	}
}

func TestTverskySmall2(t *testing.T) {
	Logger.Info().Msgf("\n    Test: running TestTverskySmall")
	clientSet := []uint64{86, 88, 134, 66, 23, 9, 83, 29, 16, 71, 144, 65, 17, 126, 150, 5, 156, 155, 140, 122, 113, 90, 10, 141, 112, 55, 39, 109, 13}
	serverSets := [][]uint64{
		{31, 51, 80},
		{80, 42, 61, 49, 62, 95},
		{158, 84, 77, 69, 150, 133, 117, 21, 47, 132, 9, 65, 87, 95, 159, 149, 71, 147, 50, 78, 154, 92, 136, 27, 41, 22, 32, 143, 52, 160, 96, 10, 130, 57, 70, 38, 102, 73, 29, 53, 144, 54, 139, 11, 107, 58, 99, 118, 40, 131, 138, 89, 105, 6, 17, 48, 128, 63, 125, 109, 45, 76, 33, 122, 80, 127, 124, 93, 104, 121, 106, 2, 55, 4, 146, 111, 140, 155, 25, 119, 36, 16, 51, 13, 35},
		{29, 153, 54, 82, 66, 68, 129, 25, 37, 163, 10, 33, 43, 137, 52, 77, 76, 93, 148, 102, 36, 90, 28, 98, 19, 7, 138, 5, 70, 134, 56, 51, 115, 149, 62, 124, 94, 107, 88, 154, 58, 127, 67, 140, 121, 17, 97, 106, 117, 86, 4, 1, 100, 166, 103, 125, 161, 39, 73, 23, 48, 143, 113, 11, 69, 13, 41, 165, 116, 12, 95, 133},
	}

	if !checkAggregatedTversky(PARAM_SIZE, clientSet, serverSets, AGGREGATION_CA_MS) {
		t.Error("Mismatch CA-MS tversky")
	}
	if !checkAggregatedTversky(PARAM_SIZE, clientSet, serverSets, AGGREGATION_X_MS) {
		t.Error("Mismatch X-MS tversky")
	}
}

func TestTverskyRandom(t *testing.T) {
	Logger.Info().Msgf("\n    Test: running TestTverskyRandom")

	sets, err := RandomDataSet(1000, 3, 100, 167)
	if err != nil {
		panic(err)
	}

	if !checkPlainTversky(PARAM_SIZE, sets[0], sets[1:]) {
		t.Error("Mismatch plain tversky")
	}
	if !checkTversky(PARAM_SIZE, sets[0], sets[1:]) {
		t.Error("Mismatch tversky")
	}
	if !checkAggregatedTversky(PARAM_SIZE, sets[0], sets[1:], AGGREGATION_CA_MS) {
		t.Error("Mismatch CA-MS tversky")
	}
	if !checkAggregatedTversky(PARAM_SIZE, sets[0], sets[1:], AGGREGATION_X_MS) {
		t.Error("Mismatch X-MS tversky")
	}
}

func TestTverskyChembl(t *testing.T) {
	Logger.Info().Msgf("\n    Test: running TestTverskyChembl")

	sets := ReadCompoundsFromFile(FPS_MINI_PATH, 1000)

	if !checkPlainTversky(PARAM_SIZE, sets[0], sets[1:]) {
		t.Error("Mismatch")
	}

	if !checkTversky(PARAM_SIZE, sets[0], sets[1:]) {
		t.Error("Mismatch")
	}
}

func TestTverskyRandomLarge(t *testing.T) {
	Logger.Info().Msgf("\n    Test: running TestTverskyRandomLarge")
	if SKIP_LONG_TESTS {
		Logger.Warn().Msgf("Skip TestTverskyRandomLarge. (takes a long time)")
		return
	}

	sets, err := RandomDataSet(16384*3+17, 3, 40, 167)
	if err != nil {
		panic(err)
	}

	if checkPlainTversky(PARAM_SIZE, sets[0], sets[1:]) == false {
		t.Error("Mismatch")
	}
}

func checkFPSMresult(t *testing.T, clientSet []uint64, serverSets [][]uint64, ans []uint64) {
	for i, v := range serverSets {
		diff := len(clientSet) - len(Intersection(clientSet, v))
		match := uint64(0)
		if diff == 0 {
			match = 1
			Logger.Debug().Msgf("Set %v is a match.", i)
		}
		if match != ans[i] {
			Logger.Warn().Msgf("Set %v is incorrect:", i)
			Logger.Warn().Msgf("  | Client: %v, server: %v", clientSet, serverSets[i])
			Logger.Warn().Msgf("  | |I|: %v, ans: %v, correct: %v",
				len(Intersection(clientSet, serverSets[i])), ans[i], match)
			t.Error("Mismatch")
		}
	}
}

func TestFPSMRandom(t *testing.T) {
	Logger.Info().Msgf("\n    Test: running TestFPSMRandom")

	sets, err := RandomDataSet(100, 3, 500, 500)
	if err != nil {
		panic(err)
	}

	clientSet, serverSets := sets[0][:3], sets[:]

	qt, err := NewQueryType(false, PSI_PSI, MATCHING_FPSM, AGGREGATION_NAIVE)
	if err != nil {
		panic(err)
	}

	_, ans := runHomoPsi(13, clientSet, serverSets, *qt, 1)

	checkFPSMresult(t, clientSet, serverSets, ans)
}

func TestFPSMRandomLong(t *testing.T) {
	Logger.Info().Msgf("\n    Test: running TestFPSMRandomLong")

	sets, err := RandomDataSet(9000, 3, 60, 1000)
	if err != nil {
		panic(err)
	}

	clientSet, serverSets := sets[0][:3], sets[:]
	serverSets[4500] = append(clientSet, 1, 8, 33, 66, 52)
	serverSets[8500] = append(clientSet, 42, 23, 55, 1)

	qt, err := NewQueryType(false, PSI_PSI, MATCHING_FPSM, AGGREGATION_NAIVE)
	if err != nil {
		panic(err)
	}

	_, ans := runHomoPsi(13, clientSet, serverSets, *qt, 8)

	checkFPSMresult(t, clientSet, serverSets, ans)
}

// Currently no test for aggregation ca-ms or x-ms
