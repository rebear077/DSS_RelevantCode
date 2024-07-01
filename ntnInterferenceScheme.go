package main

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
)

/*
	This code snippet provides an example where, under the condition that all SINR requirements are met,
	when considering whether to allocate a frequency band to a terminal,
	it is necessary to calculate the co-channel aggregate interference caused by the terminal to all terminals already using the band.
	The smaller the interference, the higher the priority for sharing.
*/

/*
	AwitTerminalList: The set of terminals waiting to share the spectrum.
*/
func (nit *NTNINTERFERENCEtpbft) TidyTransactions(leader common.NodeID, version common.Hash, nonce uint8, txs []*eles.Transaction, validOrder []byte) (result []byte, signature []byte, block *eles.Block) {
	plainTxs := make([]eles.Transaction, 0)
	for txs_i := 0; txs_i < len(txs); txs_i++ {
		if validOrder[txs_i] == byte(1) {
			tempTx := txs[txs_i]
			// if tempTx.Function ==  {
			// } else if tempTx.Function ==  {
			// } else if tempTx.Function ==  {
			// 	/* TODO
			// 	*/
			// }
			if tempTx.Function == contract.DEMO_CONTRACT_NTN_FuncOfferPriceforChannelAddr {
				channelAllocationResults := []contract.ChannelAllocation{}
				fmt.Println("The sharing process will now begin based on minimizing aggregate interference.")

				tempSharedChannelList := nit.SharedChannelList

				channel_coefficientSlice := [][]float64{}
				channel_coefficientSlice_toBytes := tempTx.Args[1]
				err := json.Unmarshal(channel_coefficientSlice_toBytes, &channel_coefficientSlice)
				if err != nil {
					continue
				}

				itfAdjSlice := [][]float64{}
				itfAdjSlice_toBytes := tempTx.Args[2]
				err = json.Unmarshal(itfAdjSlice_toBytes, &itfAdjSlice)
				if err != nil {
					continue
				}

				for awt_i := 0; awt_i < len(nit.AwitTerminalList); awt_i++ {
					fmt.Printf("Buyer AP's%s terminal %d begins matching.\n", nit.AwitTerminalList[awt_i].TerminalOwner, nit.AwitTerminalList[awt_i].TerminalID)
					chosenTerminal := nit.AwitTerminalList[awt_i]
					hp := &Heap1{}
					for tmpCh := 0; tmpCh < len(tempSharedChannelList); tmpCh++ {
						if channel2Frequency[tempSharedChannelList[tmpCh].ChannelID] >= chosenTerminal.LowerChannelSupported && channel2Frequency[tempSharedChannelList[tmpCh].ChannelID] <= chosenTerminal.UpperChannelSupported {
							sumCausedInterference := 0.0
							// The terminals already present on this channel.
							existTerminals := append([]contract.TerminalInfo{}, nit.ChannelOccupancy[tempSharedChannelList[tmpCh]]...)
							// The newly added terminal and the AP to which the terminal belongs.
							apIndex1 := nit.APInfo[chosenTerminal.TerminalOwner].Index
							for eti := 0; eti < len(existTerminals); eti++ {
								dis := calculateEuclideanDistance3D(
									TerminalPos[existTerminals[eti].TerminalID][0], TerminalPos[existTerminals[eti].TerminalID][1], TerminalPos[existTerminals[eti].TerminalID][2],
									ApPos[apIndex1-1][0], ApPos[apIndex1-1][1], ApPos[apIndex1-1][2])
								exp1 := channel_coefficientSlice[apIndex1-1][existTerminals[eti].TerminalID]
								adj1 := itfAdjSlice[apIndex1-1][existTerminals[eti].TerminalID]
								sumCausedInterference += CaculateInterference(dis, float64(chosenTerminal.Power), exp1, adj1)
							}
							heap.Push(hp, tempSumInterferenceInfo{
								Channel:               tempSharedChannelList[tmpCh],
								SumInterferenceCaused: sumCausedInterference,
							})
						} else {
							continue
						}
					}
					for hp.Len() > 0 {
						x := heap.Pop(hp)
						tmpSumInterferenceInfo := x.(tempSumInterferenceInfo)
						tmpChannel := tmpSumInterferenceInfo.Channel
						tempTerminals := append([]contract.TerminalInfo{chosenTerminal}, nit.ChannelOccupancy[tmpChannel]...)
						sinrCheck := 0
						for tt := 0; tt < len(tempTerminals); tt++ {
							checkTerminal := tempTerminals[tt]
							apIndex1 := nit.APInfo[checkTerminal.TerminalOwner].Index
							dis := calculateEuclideanDistance3D(
								TerminalPos[checkTerminal.TerminalID][0], TerminalPos[checkTerminal.TerminalID][1], TerminalPos[checkTerminal.TerminalID][2],
								ApPos[apIndex1-1][0], ApPos[apIndex1-1][1], ApPos[apIndex1-1][2])
							exp1 := channel_coefficientSlice[apIndex1-1][checkTerminal.TerminalID]
							S := float64(checkTerminal.Power) * exp1 * math.Pow(1.0/dis, Alpha)
							SINR_Interference := N0
							for q := 0; q < len(tempTerminals); q++ {
								if q == tt {
									continue
								}
								SINR_InterferenceTerminal := tempTerminals[q]

								apIndex2 := nit.APInfo[SINR_InterferenceTerminal.TerminalOwner].Index
								d := calculateEuclideanDistance3D(
									TerminalPos[checkTerminal.TerminalID][0], TerminalPos[checkTerminal.TerminalID][1], TerminalPos[checkTerminal.TerminalID][2],
									ApPos[apIndex2-1][0], ApPos[apIndex2-1][1], ApPos[apIndex2-1][2])
								exp2 := channel_coefficientSlice[apIndex2-1][checkTerminal.TerminalID]
								adj2 := itfAdjSlice[apIndex2-1][checkTerminal.TerminalID]
								SINR_Interference += CaculateInterference(d, float64(SINR_InterferenceTerminal.Power), exp2, adj2)
								if (S / SINR_Interference) < SINR_th {
									// Already violating SINR requirements.
									sinrCheck = 1
									break
								}
							}
							if sinrCheck == 1 {
								break
							}
						}
						if sinrCheck == 1 {
							continue
						}
						// Pass the SINR condition check.
						nit.ChannelOccupancy[tmpChannel] = append(nit.ChannelOccupancy[tmpChannel], chosenTerminal)
						sellerAP := tmpChannel.OwnerAP
						buyerAP := chosenTerminal.TerminalOwner
						round := strconv.Itoa(nit.NTNRound)

						tmp := contract.ChannelAllocation{
							Round:    round,
							SellerAP: sellerAP,
							BuyerAP:  buyerAP,
							Channel:  tmpChannel.ChannelID,
							Terminal: chosenTerminal,
						}
						channelAllocationResults = append(channelAllocationResults, tmp)
						nit.TransactionPair[SellerBuyerPair{sellerAP, buyerAP}] = append(nit.TransactionPair[SellerBuyerPair{sellerAP, buyerAP}], tmp)
						res1 := nit.APInfo[buyerAP]
						res1.ChannelOnLoan = append(res1.ChannelOnLoan, tmpChannel)
						nit.APInfo[buyerAP] = res1

						res2 := nit.APInfo[sellerAP]
						res2.ChannelLending = append(res2.ChannelLending, tmpChannel)
						nit.APInfo[sellerAP] = res2
						break
					}
				}
				fmt.Println("DoneDoneDone")

				resResult := []contract.RoundResultRecord{}
				for a_i := 0; a_i < len(nit.APIDList); a_i++ {
					ap := nit.APInfo[nit.APIDList[a_i]]
					round := strconv.Itoa(nit.NTNRound)
					r := contract.RoundResultRecord{
						Round:          round,
						APID:           ap.APID,
						ChannelOnLoan:  ap.ChannelOnLoan,
						ChannelLending: ap.ChannelLending,
					}
					resResult = append(resResult, r)
				}

				channelAllocationResults_toBytes, _ := json.Marshal(channelAllocationResults)
				resResult_toBytes, _ := json.Marshal(resResult)

				round := strconv.Itoa(nit.NTNRound)
				txArgs := [][]byte{[]byte(round), channelAllocationResults_toBytes}
				tmpTx := nit.processChannelAllocationTx(version, nonce, txArgs)
				plainTxs = append(plainTxs, *tmpTx)

				txArgs2 := [][]byte{[]byte(round), resResult_toBytes}
				tmpTx2 := nit.processResultRecordTx(version, nonce, txArgs2)
				plainTxs = append(plainTxs, *tmpTx2)

				// rest AP's status
				resetAPInfo := []contract.APEntity{}
				for a_i := 0; a_i < len(nit.APIDList); a_i++ {
					ap := nit.APInfo[nit.APIDList[a_i]]
					resetAPInfo = append(resetAPInfo, ap)
				}

				// tx to rest AP's status
				rstAPbyte, _ := json.Marshal(resetAPInfo)
				txArgs3 := [][]byte{rstAPbyte}
				tempTx := nit.processResetStatusTx(version, nonce, txArgs3)
				plainTxs = append(plainTxs, *tempTx)

				// Reset the following variables to their initial state.
				nit.AwitTerminalList = []contract.TerminalInfo{}
				nit.ChannelOccupancy = make(map[contract.ChannelInfo][]contract.TerminalInfo)
				nit.SharedChannelList = make([]contract.ChannelInfo, 0)
				nit.TransactionPair = make(map[SellerBuyerPair][]contract.ChannelAllocation)

				for ra_i := 0; ra_i < len(nit.RawAPInfoList); ra_i++ {
					apid := nit.RawAPInfoList[ra_i].APID
					nit.APInfo[apid] = nit.RawAPInfoList[ra_i]
				}

				nit.NTNRound++
				time.Sleep(2 * time.Second)
			}
		}
	}

	leaderAddress, err := crypto.NodeIDtoAddress(leader) //Calculate the address based on the NodeID.
	if err != nil {
		loglogrus.Log.Warnf("[TPBFT Consensus] NTNInterferencetpbft failed: Couldn't tidy Transactions into candidateBlock, it's not to assert NodeID into Address!\n")
		return nil, nil, nil
	}

	//构建区块
	candidateBlock := &eles.Block{
		BlockID:      common.Hash{},
		Subnet:       []byte(nit.consensusPromoter.SelfNode.NetID), //The subnet ID of the current node
		Leader:       leaderAddress,                                //The address of the leader node
		Version:      version,
		Nonce:        nonce,
		Transactions: plainTxs,                        //Set composed of valid transactions
		SubnetVotes:  make([]eles.SubNetSignature, 0), //Used to store the digital signatures of participants in the consensus process.
		PrevBlock:    common.Hash{},                   //The hash of the previous block
		CheckRoot:    common.Hash{},
		Receipt:      eles.NullReceipt,
		LeaderVotes:  make([]eles.LeaderSignature, 0), //When a block is published to the blockchain, it needs to record the signatures of all leader nodes involved in the consensus process (upper-layer consensus).
	}

	blockID, err := candidateBlock.ComputeBlockID() //Use the hash computed from the initialized block as the block ID.
	if err != nil {
		loglogrus.Log.Warnf("[TPBFT Consensus] NTNInterferencetpbft failed: Couldn't tidy Transactions into candidateBlock, Compute BlockID is failed, err:%v\n", err)
		return nil, nil, nil
	}
	result = blockID[:]
	signature, err = crypto.Sign(result, nit.consensusPromoter.prvKey) //Generate a digital signature using the block ID.
	if err != nil {
		loglogrus.Log.Warnf("[TPBFT Consensus] NTNInterferencetpbft failed: Couldn't tidy Transactions into candidateBlock, Generate digtal signature is failed, err:%v\n", err)
		return nil, nil, nil
	}
	loglogrus.Log.Infof("[TPBFT Consensus] NTNInterferencetpbft: Leader (%x) Tidy candidateBlock successfully, blockID:%x, txSum:%d\n", leaderAddress, candidateBlock.BlockID, len(candidateBlock.Transactions))
	return result, signature, candidateBlock
}
