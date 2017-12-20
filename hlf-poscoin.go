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

	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"crypto/sha256"
)

var logger = shim.NewLogger("mylogger")

const oneDayUnixTime int = 86400000000000
const oneYearUnixTime int = 31536000000000000

type AerialCC struct {
	name string
	symbol string
	decimals int

	chainStartTime int
	chainStartBlockNumber int
	stakeStartTime int
	stakeMinAge int
	stakeMaxAge int
	maxMineProofOfStake int

	totalSupply int
	maxTotalSupply int
	totalInitialSupply int

}

type TransferInStruct struct {
	Address string "json:address"
	Amount int "json:amount"
	Time int "json:time"
}
type transferIns []TransferInStruct


// Called to initialize the chaincode
func (t *AerialCC) Init(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var err error

	logger.Info("Starting Initializing the Chaincode")

	if len(args) < 12 {
		logger.Error("Invalid number of arguments")
		return nil, errors.New("Invalid number of arguments")
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
	t.maxMineProofOfStake = 100000000000000000

	t.totalSupply = 100
	t.maxTotalSupply = 21000000
	t.totalInitialSupply = 100
	logger.Info("Successfully Initialized the AerialCC")

	return nil, nil

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

func (t *AerialCC) increaseTotalSupply(stub shim.ChaincodeStubInterface, reward int) ([]byte, error) {
	t.totalSupply = t.totalSupply + reward
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
	src, _ = strconv.Atoi(src)
	dst, _ = strconv.Atoi(dst)
	src = src - X
	dst = dst + X
	logger.Info("srcAmount = %d, dstAmount = %d\n", src, dst)

	err = stub.PutState(args[0], []byte(strconv.Itoa(src)))
	if err != nil {
		logger.Error("failed to write the state for src!")
		return nil, err
	}

	err = stub.PutState(args[1], []byte(strconv.Itoa(dst)))
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

	val, err := stub.GetState(stub, args[0])
	if err != nil {
		return nil, err
	}
	logger.Info("Query Response: %d\n", val)
	return val, nil
}

func MinePoS(stub shim.ChaincodeStubInterface, args []string) (bool,error) {

	//canPoSMint
	src, err := stub.GetState(stub, args[0])
	if err != nil {
		return false, err
	}

	st := string(append(args[0],"transferIn"))
	transferinsID := sha256.New()
	transferinsID.Write([]byte (st))
	transferIns, err := stub.GetState(transferinsID.Sum(nil))
	var um transferIns
	err = json.Unmarshal(transferIns, &um)

	if err != nil {
		return false, err
	}

	if len(transferIns) <= 0 {
		return false, err
	}

	reward := t.getProofOfStakeReward(stub, args[0])
	if reward <= 0 {
		return false, err
	}

	newTS, err := t.increaseTotalSupply(reward)
	if err != nil {
		fmt.Printf("IncreaseTotalSupply Failed: %s", err)
		return false, err
	}

	src = src + reward
	err = stub.PutState(args[0], []byte(strconv.Itoa(src)))
	if err != nil {
		return false, err
	}
	fmt.Println("sup!?")
	um = nil
	var temp_tin TransferInStruct
	temp_tin.Address = param.PartySrc
	temp_tin.Amount = src+reward
	temp_tin.Time = time.Now().Unix()

	um = append(um, temp_tin)
	um, err = json.Marshal(&um)
	if err != nil {
		return false, err
	}
	stub.PutState(transferinsID.Sum(), um)

	return true, nil
}

func getProofOfStakeReward(stub shim.ChaincodeStubInterface, args []string) (int, bool) {

	now := time.Now().Unix()
	if now <= t.stakeStartTime || stakeStartTime <= 0 {
		return 0,false
	}

	_coinAge = getCoinAge(stub, now, args)
	if _coinAge <= 0 {
		return 0, false
	}

	var interest int
	interest = t.maxMintProofOfStake
	if (now - t.stakeStartTime) / oneYearUnixTime == 0 {
		interest = (770 * t.maxMintProofOfStake) / 100
	} else if (now - t.stakeStartTime) / oneYearUnixTime == 1 {
		interest = (435 * maxMintProofOfStake) / 100
	}

	return (_coinAge * interest) / (365* (10**t.decimals)), true

}

func getCoinAge(stub shim.ChaincodeStubInterface, now time, args []string) (int, bool) {

	st := string(append(args[0],"transferIn"))
	transferinsID := sha256.New()
	transferinsID.Write([]byte (st))
	transferIns, err := stub.GetState(transferinsID.Sum(nil))
	var um transferIns
	err = json.Unmarshal(transferIns, &um)

	if err != nil {
		return 0, false
	}

	if len(transferIns) <= 0 {
		return 0, false
	}

	var _coinAge int
	for i := 0; i < len(transferIns); i++ {
		if now.Unix() < (transferIns[i].Time + t.stakeMinAge){
			continue
		}
		var nCoinSeconds int
		nCoinSeconds = now.Unix - transferIns[i].Time
		if nCoinSeconds > t.stakeMaxAge {
			nCoinSeconds = t.stakeMaxAge
		}
		_coinAge = _coinAge + transferIns[i].Amount * (nCoinSeconds / 86400*(10**9))
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
