package psm

import (
	"fmt"
	"time"
)

type BenchData struct {
	SetNum       int
	RespSize     int
	QuerySize    int
	PreProcess   float64
	Query        float64
	Response     float64
	Evaluation   float64
	QueryMarshal float64
	RespMarshal  float64
	KeyGen       float64
	Latency      float64
}

func BenchHomoPSI(pp *PSIParams, sets [][]uint64, queryType QueryType) BenchData {
	startTime := time.Now()
	clinetSet := sets[0]
	serverSets := sets[1:]
	dataGenTime := time.Now()

	// DescribeParams(pp.params)

	cl := NewClient(pp)
	keyGenTime := time.Now()
	sv, err := NewServer(pp, serverSets)
	if err != nil {
		panic(err)
	}
	clKey := cl.GetKey()
	paramTime := time.Now()

	query, err := cl.Query(clinetSet, queryType)
	if err != nil {
		panic(err)
	}
	queryTime := time.Now()
	resp, err := sv.Respond(query, clKey)
	if err != nil {
		panic(err)
	}
	respTime := time.Now()
	ans := cl.EvalResponse(clinetSet, query, resp)
	endTime := time.Now()

	queryMarshalled, _ := query.MarshalBinary()
	queryMarshalTime := time.Now()
	respMarshalled, _ := resp.MarshalBinary()
	respMarshalTime := time.Now()

	clTotalTime := queryTime.Sub(paramTime) + endTime.Sub(respTime) + queryMarshalTime.Sub(endTime)
	svTotalTime := respTime.Sub(queryTime) + respMarshalTime.Sub(queryMarshalTime)

	// fmt.Println("Client set: ", clinetSet)
	// fmt.Println("Number of server sets: ", len(serverSets))
	_ = ans

	// toMb := 1024 * 1024
	toMb64 := uint64(1024 * 1024)

	fmt.Println("\n***************************************************")
	fmt.Printf("* Computation\n")
	fmt.Printf("* #server sets:               %v\n", len(serverSets))
	fmt.Printf("* Random set gen:             %v\n", dataGenTime.Sub(startTime))
	fmt.Printf("* Query:                      %v\n", queryTime.Sub(paramTime))
	fmt.Printf("* Response:                   %v\n", respTime.Sub(queryTime))
	fmt.Printf("* Evaluation:                 %v\n", endTime.Sub(respTime))
	fmt.Printf("* Query Marshal:              %v\n", queryMarshalTime.Sub(endTime))
	fmt.Printf("* Resp Marshal:               %v\n", respMarshalTime.Sub(queryMarshalTime))
	fmt.Printf("* Client total  =>  %v\n", clTotalTime)
	fmt.Printf("* Server total  =>  %v\n", svTotalTime)
	fmt.Println("***************************************************")
	fmt.Printf("* Communication\n")
	fmt.Printf("* Query size:                 %v KB\n", len(queryMarshalled)/1024)
	fmt.Printf("* Response size:              %v KB\n", len(respMarshalled)/1024)
	fmt.Println("***************************************************")
	fmt.Printf("* Key generation\n")
	fmt.Printf("* Time:                       %v\n", keyGenTime.Sub(dataGenTime))
	fmt.Printf("* Public key size:            %v KB\n", clKey.pk.PublicKey.GetDataLen(true)/1024)
	fmt.Printf("* Relin key size:             %v MB\n", clKey.evk.Rlk.GetDataLen(true)/toMb64)
	fmt.Printf("* Rotate key size:            %v MB\n", clKey.evk.Rtks.GetDataLen(true)/toMb64)
	fmt.Println("***************************************************")

	fmt.Println("Answer: ", ans)

	return BenchData{
		SetNum:       len(serverSets),
		RespSize:     len(respMarshalled),
		QuerySize:    len(queryMarshalled),
		PreProcess:   paramTime.Sub(keyGenTime).Seconds(),
		Query:        queryTime.Sub(paramTime).Seconds(),
		Response:     respTime.Sub(queryTime).Seconds(),
		Evaluation:   endTime.Sub(respTime).Seconds(),
		QueryMarshal: queryMarshalTime.Sub(endTime).Seconds(),
		RespMarshal:  respMarshalTime.Sub(queryMarshalTime).Seconds(),
		KeyGen:       keyGenTime.Sub(dataGenTime).Seconds(),
		Latency:      respMarshalTime.Sub(paramTime).Seconds(),
	}
}

func APISample() {
	var serverSets [][]uint64
	var clientSet []uint64

	bfvParams := GetBFVParam(15)       // BFV params with N=2^15
	pp := NewPSIParams(bfvParams, 128) // A framework params with 128 bit se

	queryType, err := NewQueryType(true, PSI_CA, MATCHING_TVERSKY, AGGREGATION_NAIVE)
	if err != nil {
		panic(err)
	}

	// Setup phase
	cl := NewClient(pp)
	sv, err := NewServer(pp, serverSets)
	if err != nil {
		panic(err)
	}
	clKey := cl.GetKey()

	// Query
	query, err := cl.Query(clientSet, *queryType)
	if err != nil {
		panic(err)
	}
	resp, err := sv.Respond(query, clKey)
	if err != nil {
		panic(err)
	}
	ans := cl.EvalResponse(clientSet, query, resp)
	_ = ans
}
