/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at
  http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"math"

	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"crypto/sha256"
)

var logger = shim.NewLogger("mylogger")

const oneDayUnixTime int64 = 86400000000000
const oneYearUnixTime int64 = 31536000000000000

type AerialCC struct {
	name string
	symbol string
	decimals int

	chainStartTime int64
	chainStartBlockNumber int
	stakeStartTime int64
	stakeMinAge int64
	stakeMaxAge int64
	maxMintProofOfStake int

	totalSupply int
	maxTotalSupply int
	totalInitialSupply int

}

type TransferInStruct struct {
	Address string "json:address"
	Amount int64 "json:amount"
	Time int64 "json:time"
}

type transferIns []TransferInStruct

type aerialResponse struct {
    // A status code that should follow the HTTP status codes.
    Status int32 `protobuf:"varint,1,opt,name=status" json:"status,omitempty"`
    // A message associated with the response code.
    Message string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
    // A payload that can be used to include metadata with this response.
    Payload []byte `protobuf:"bytes,3,opt,name=payload,proto3" json:"payload,omitempty"`
}

// Called to initialize the chaincode

func (t *AerialCC) Init(stub shim.ChaincodeStubInterface) peer.Response {

	args := stub.GetStringArgs()

	logger.Info("Starting Initializing the Chaincode")

	var resp peer.Response

	if len(args) < 12 {
		logger.Error("Invalid number of arguments")
		resp.Status = 1
		resp.Message = "Invalid number of args"
		resp.Payload = nil
		return resp
	}

	/**
	 0:name
	 1:symbol
	 2:decimals
	 3:chainstarttime
	 4:stakestarttime
	 5:chainStartBlockNumber
	 6:stakeMinAge
	 7:stakeMaxAge
	 8:maxMineProofOfStake
	 9:totalSupply
	 10:maxTotalSupply
	 11:totalInitialSupply
	 **/

	fmt.Println("args[0] = %s", args[0])

 /**
	t.name = args[0]
	t.symbol = args[1]
	t.decimals = strconv.Atoi(args[2])
	//Timings
	chainStartTime := strconv.Atoi(args[3])
	stakeStartTime := strconv.Atoi(args[4])
	const shortForm = "2006-Jan-02"
	f, _ := time.Parse(shortForm, chainStartTime)
	g, _ := time.Parse(shortForm, stakeStartTime)
	t.chainStartTime = int32(f.Unix())
	t.stakeStartTime = int32(g.Unix())

	t.chainStartBlockNumber = strconv.Atoi(args[5])
	t.stakeMinAge = strconv.Atoi(args[6])*oneDayUnixTime
	t.stakeMaxAge = strconv.Atoi(args[7])*oneDayUnixTime
	t.maxMineProofOfStake = strconv.Atoi(args[8])

	t.totalSupply = strconv.Atoi(args[9])
	t.maxTotalSupply = strconv.Atoi(args[10])
	t.totalInitialSupply = strconv.Atoi(args[11])
**/

	t.name = "cryptorial"
	t.symbol = "cri"
	t.decimals = 18
	//Timings

	t.stakeMinAge = 3*oneDayUnixTime
	t.stakeMaxAge = 90*oneDayUnixTime
	t.maxMintProofOfStake = 100000000000000000

	t.totalSupply = 100
	t.maxTotalSupply = 21000000
	t.totalInitialSupply = 100
	logger.Info("Successfully Initialized the AerialCC")

	resp.Status = 2
	resp.Message = "Vaild everything"
	resp.Payload = nil
	return resp

}
func (t *AerialCC) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	var resp peer.Response
	resp.Status = 1
	resp.Message = "invoked"
	resp.Payload = nil
	return resp
}
func (t *AerialCC) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "MakePayment" {
		return MakePayment(stub, args)
	} else if function == "DeleteAccount" {
		return DeleteAccount(stub, args)
		} else if function == "CheckBalance" {
			return CheckBalance(stub, args)
		}
	return nil, nil
}

func (t *AerialCC) increaseTotalSupply(stub shim.ChaincodeStubInterface, reward int64) ([]byte, error) {
	t.totalSupply = t.totalSupply + int(reward)
	return nil, nil
}

// Transaction makes payment of X units from A to B
func MakePayment(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var err error

	src, err := stub.GetState(args[0])
	if err != nil {
		logger.Error("partySrc is missing!")
		return nil, err
	}

	dst, err := stub.GetState(args[1])
	if err != nil {
		logger.Error("partyDst is missing!")
		return nil, err
	}

	X, _ := strconv.Atoi(args[2])
	src_str, _ := strconv.Atoi(string(src))
	dst_str, _ := strconv.Atoi(string(dst))
	src = []byte(strconv.Itoa(src_str - X))
	dst = []byte(strconv.Itoa(dst_str + X))
	logger.Info("srcAmount = %d, dstAmount = %d\n", src, dst)

	err = stub.PutState(args[0], src)
	if err != nil {
		logger.Error("failed to write the state for src!")
		return nil, err
	}

	err = stub.PutState(args[1], dst)
	if err != nil {
		logger.Error("failed to write the state for dst!")
		return nil, err
	}

	return nil, nil
}

// Deletes an entity from state
func DeleteAccount(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	err := stub.DelState(args[0])
	if err != nil {
		logger.Error("Failed to delete state!")
		return nil, errors.New("Failed to delete state")
	}

	return nil, nil

}

// Query callback representing the query of a chaincode
func CheckBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	val, err := stub.GetState(args[0])
	if err != nil {
		return nil, err
	}
	logger.Info("Query Response: %d\n", val)
	return val, nil
}

func (t *AerialCC) MinePoS(stub shim.ChaincodeStubInterface, args []string) (bool,error) {

	//canPoSMint
	src, err := stub.GetState(args[0])
	if err != nil {
		return false, err
	}

	st := string(args[0]) + "transferIn"
	transferinsID := sha256.New()
	transferinsID.Write([]byte (st))
	_transferIns, err := stub.GetState(string(transferinsID.Sum(nil)))
	var um []TransferInStruct
	err = json.Unmarshal(_transferIns, &um)

	if err != nil {
		return false, err
	}

	if len(um) <= 0 {
		return false, err
	}

	reward, _ := t.getProofOfStakeReward(stub, args[0])
	if reward <= 0 {
		return false, err
	}

	newTS, err := t.increaseTotalSupply(stub, reward)
	if err != nil {
		fmt.Printf("IncreaseTotalSupply Failed: %s", err)
		return false, err
	}
	fmt.Printf("Total Supply Increased to: %s", newTS)
	src_integer, _ := strconv.Atoi(string(src))
	src = []byte(strconv.Itoa(src_integer + int(reward)))
	err = stub.PutState(args[0], src)
	if err != nil {
		return false, err
	}
	fmt.Println("sup!?")
	//um := nil
	var um_new []TransferInStruct
	var temp_tin TransferInStruct
	temp_tin.Address = args[0]
	temp_tin.Amount = int64(src_integer + int(reward))
	temp_tin.Time = time.Now().Unix()

	um = append(um_new, temp_tin)
	um_b, err := json.Marshal(&um)
	if err != nil {
		return false, err
	}
	stub.PutState(string(transferinsID.Sum(nil)), um_b)

	return true, nil
}

func (t *AerialCC) getProofOfStakeReward(stub shim.ChaincodeStubInterface, address string) (int64, bool) {

	now := time.Now().Unix()
	if now <= t.stakeStartTime || t.stakeStartTime <= 0 {
		return 0,false
	}

	_coinAge, _ := t.getCoinAge(stub, now, address)
	if _coinAge <= 0 {
		return 0, false
	}

	var interest int
	interest = t.maxMintProofOfStake
	if (int64(now) - t.stakeStartTime) / oneYearUnixTime == 0 {
		interest = (770 * t.maxMintProofOfStake) / 100
	} else if (now - t.stakeStartTime) / oneYearUnixTime == 1 {
		interest = (435 * t.maxMintProofOfStake) / 100
	}

	return int64(float64(_coinAge * int64(interest)) / (365* (math.Pow(10,float64(t.decimals))))), true

}

func (t *AerialCC) getCoinAge(stub shim.ChaincodeStubInterface, now int64, address string) (int64, bool) {

	st := address + "transferIn"
	transferinsID := sha256.New()
	transferinsID.Write([]byte (st))
	transferIns_state, err := stub.GetState(string(transferinsID.Sum(nil)))
	var um []TransferInStruct
	err = json.Unmarshal(transferIns_state, &um)

	if err != nil {
		return 0, false
	}

	if len(um) <= 0 {
		return 0, false
	}

	var _coinAge int64
	for i := 0; i < len(um); i++ {
		if now < (um[i].Time + t.stakeMinAge){
			continue
		}
		var nCoinSeconds int64
		nCoinSeconds = now - um[i].Time
		if nCoinSeconds > t.stakeMaxAge {
			nCoinSeconds = t.stakeMaxAge
		}
		_coinAge = _coinAge + um[i].Amount * (nCoinSeconds / int64(86400*(math.Pow(10,9))))
	}
	return _coinAge, true
}

func main() {

	lld, _ := shim.LogLevel("DEBUG")
	fmt.Println(lld)

	logger.SetLevel(lld)
	fmt.Println(logger.IsEnabledFor(lld))

	err := shim.Start(new(AerialCC))
	if err != nil {
		logger.Error("Could not start AerialCC")
	} else {
		logger.Info("AerialCC successfully started")
	}

}
