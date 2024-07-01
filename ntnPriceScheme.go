package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"
)

/*
	This code snippet provides an example where,
	under the condition that all SINR requirements are met,
	the buyer with the highest bid is prioritized for spectrum sharing.
*/

/*
	AwitTerminalList: The set of terminals waiting to share the spectrum.
*/

// TidyTransactions
func (npt *NTNPRICEtpbft) TidyTransactions(leader common.NodeID, version common.Hash, nonce uint8, txs []*eles.Transaction, validOrder []byte) (result []byte, signature []byte, block *eles.Block) {
	plainTxs := make([]eles.Transaction, 0) //Store all valid transactions.

	for txs_i := 0; txs_i < len(txs); txs_i++ {
		// Iterate through the entire transaction set and extract only the valid transactions
		// (where the corresponding bit in validOrder is equal to 1).
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
				fmt.Println("The sharing process will now begin based on the bidding rules.")
				// The highest bidder gets priority in the transaction.
				sort.SliceStable(npt.AwitTerminalList, func(i, j int) bool {
					return npt.AwitTerminalList[i].Price > npt.AwitTerminalList[j].Price
				})

				tempSharedChannelList := npt.SharedChannelList
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

				for awt_i := 0; awt_i < len(npt.AwitTerminalList); awt_i++ {
					fmt.Printf("Buyer AP's%s terminal %d begins matching.\n", npt.AwitTerminalList[awt_i].TerminalOwner, npt.AwitTerminalList[awt_i].TerminalID)
					chosenTerminal := npt.AwitTerminalList[awt_i]
					sinrCheck := 0
					// Traverse the currently available channels for sharing.
					for tmpCh := 0; tmpCh < len(tempSharedChannelList); tmpCh++ {
						// First, check if the channel bandwidth is within the terminal's supported range.
						if channel2Frequency[tempSharedChannelList[tmpCh].ChannelID] >= chosenTerminal.LowerChannelSupported && channel2Frequency[tempSharedChannelList[tmpCh].ChannelID] <= chosenTerminal.UpperChannelSupported {
							sinrCheck = 2
							// tempTerminals includes the terminal to be checked for inclusion, as well as the terminals already utilizing the frequency band.
							tempTerminals := append([]contract.TerminalInfo{chosenTerminal}, npt.ChannelOccupancy[tempSharedChannelList[tmpCh]]...)

							// If this terminal is added to the frequency band, the SINR condition of this terminal itself must be satisfied,
							// and the SINR conditions of other terminals already using the frequency band must not be compromised.
							for p := 0; p < len(tempTerminals); p++ {
								checkTerminal := tempTerminals[p]
								apIndex1 := npt.APInfo[checkTerminal.TerminalOwner].Index
								dis := calculateEuclideanDistance3D(
									TerminalPos[checkTerminal.TerminalID][0], TerminalPos[checkTerminal.TerminalID][1], TerminalPos[checkTerminal.TerminalID][2],
									ApPos[apIndex1-1][0], ApPos[apIndex1-1][1], ApPos[apIndex1-1][2])
								exp1 := channel_coefficientSlice[apIndex1-1][checkTerminal.TerminalID]
								S := float64(checkTerminal.Power) * exp1 * math.Pow(1.0/dis, Alpha)
								SINR_Interference := N0
								for q := 0; q < len(tempTerminals); q++ {
									if q == p {
										continue
									}
									SINR_InterferenceTerminal := tempTerminals[q]
									apIndex2 := npt.APInfo[SINR_InterferenceTerminal.TerminalOwner].Index

									d := calculateEuclideanDistance3D(
										TerminalPos[checkTerminal.TerminalID][0], TerminalPos[checkTerminal.TerminalID][1], TerminalPos[checkTerminal.TerminalID][2],
										ApPos[apIndex2-1][0], ApPos[apIndex2-1][1], ApPos[apIndex2-1][2])
									exp2 := channel_coefficientSlice[apIndex2-1][checkTerminal.TerminalID]
									adj2 := itfAdjSlice[apIndex2-1][checkTerminal.TerminalID]
									SINR_Interference += CaculateInterference(d, float64(SINR_InterferenceTerminal.Power), exp2, adj2)
									if (S / SINR_Interference) < SINR_th {
										sinrCheck = 1
										break
									}
								}
								if sinrCheck == 1 {
									break
								}
							}
							fmt.Println("----")
							if sinrCheck == 2 {
								// Prove that the check has been completed and that all SINR requirements are met.
								// Add the AP to the usage of the channel.
								npt.ChannelOccupancy[tempSharedChannelList[tmpCh]] = append(npt.ChannelOccupancy[tempSharedChannelList[tmpCh]], chosenTerminal)
								sellerAP := tempSharedChannelList[tmpCh].OwnerAP
								buyerAP := chosenTerminal.TerminalOwner

								round := strconv.Itoa(npt.NTNRound)
								tmp := contract.ChannelAllocation{
									Round:    round,
									SellerAP: sellerAP,
									BuyerAP:  buyerAP,
									Channel:  tempSharedChannelList[tmpCh].ChannelID,
									Terminal: chosenTerminal,
								}
								channelAllocationResults = append(channelAllocationResults, tmp)

								npt.TransactionPair[SellerBuyerPair{sellerAP, buyerAP}] = append(npt.TransactionPair[SellerBuyerPair{sellerAP, buyerAP}], tmp)

								res1 := npt.APInfo[buyerAP]
								res1.ChannelOnLoan = append(res1.ChannelOnLoan, tempSharedChannelList[tmpCh])
								npt.APInfo[buyerAP] = res1

								res2 := npt.APInfo[sellerAP]
								res2.ChannelLending = append(res2.ChannelLending, tempSharedChannelList[tmpCh])
								npt.APInfo[sellerAP] = res2

								temp1 := append([]contract.ChannelInfo{}, tempSharedChannelList[:tmpCh]...)
								temp2 := append([]contract.ChannelInfo{}, tempSharedChannelList[tmpCh+1:]...)
								temp1 = append(temp1, temp2...)
								temp1 = append(temp1, tempSharedChannelList[tmpCh])
								tempSharedChannelList = make([]contract.ChannelInfo, 0)
								tempSharedChannelList = append(tempSharedChannelList, temp1...)
								break
							}
						} else {
							continue
						}
					}
				}
				fmt.Println("Done")

				resResult := []contract.RoundResultRecord{}
				for a_i := 0; a_i < len(npt.APIDList); a_i++ {
					ap := npt.APInfo[npt.APIDList[a_i]]
					round := strconv.Itoa(npt.NTNRound)
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

				round := strconv.Itoa(npt.NTNRound)
				txArgs := [][]byte{[]byte(round), channelAllocationResults_toBytes}
				tmpTx := npt.processChannelAllocationTx(version, nonce, txArgs)
				plainTxs = append(plainTxs, *tmpTx)

				txArgs2 := [][]byte{[]byte(round), resResult_toBytes}
				tmpTx2 := npt.processResultRecordTx(version, nonce, txArgs2)
				plainTxs = append(plainTxs, *tmpTx2)

				// tx to rest AP's status
				resetAPInfo := []contract.APEntity{}
				for a_i := 0; a_i < len(npt.APIDList); a_i++ {
					ap := npt.APInfo[npt.APIDList[a_i]]
					resetAPInfo = append(resetAPInfo, ap)
				}

				rstAPbyte, _ := json.Marshal(resetAPInfo)
				txArgs3 := [][]byte{rstAPbyte}
				tmpTx3 := npt.processResetStatusTx(version, nonce, txArgs3)
				plainTxs = append(plainTxs, *tmpTx3)

				// Reset the following variables to their initial state.
				npt.AwitTerminalList = []contract.TerminalInfo{}
				npt.ChannelOccupancy = make(map[contract.ChannelInfo][]contract.TerminalInfo)
				npt.SharedChannelList = make([]contract.ChannelInfo, 0)
				npt.TransactionPair = make(map[SellerBuyerPair][]contract.ChannelAllocation)

				for ra_i := 0; ra_i < len(npt.RawAPInfoList); ra_i++ {
					apid := npt.RawAPInfoList[ra_i].APID
					npt.APInfo[apid] = npt.RawAPInfoList[ra_i]
				}

				npt.NTNRound++
				time.Sleep(2 * time.Second)
			}
		}
	}

	leaderAddress, err := crypto.NodeIDtoAddress(leader) //Calculate the address based on the NodeID.
	if err != nil {
		loglogrus.Log.Warnf("[TPBFT Consensus] NTNPRICEtpbft failed: Couldn't tidy Transactions into candidateBlock, it's not to assert NodeID into Address!\n")
		return nil, nil, nil
	}

	//构建区块
	candidateBlock := &eles.Block{
		BlockID:      common.Hash{},
		Subnet:       []byte(npt.consensusPromoter.SelfNode.NetID), //The subnet ID of the current node
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

	time.Sleep(2 * time.Second)

	blockID, err := candidateBlock.ComputeBlockID() //Use the hash computed from the initialized block as the block ID.
	if err != nil {
		loglogrus.Log.Warnf("[TPBFT Consensus] NTNPRICEtpbft failed: Couldn't tidy Transactions into candidateBlock, Compute BlockID is failed, err:%v\n", err)
		return nil, nil, nil
	}
	result = blockID[:]
	signature, err = crypto.Sign(result, npt.consensusPromoter.prvKey) //Generate a digital signature using the block ID.
	if err != nil {
		loglogrus.Log.Warnf("[TPBFT Consensus] NTNPRICEtpbft failed: Couldn't tidy Transactions into candidateBlock, Generate digtal signature is failed, err:%v\n", err)
		return nil, nil, nil
	}

	loglogrus.Log.Infof("[TPBFT Consensus] NTNPRICEtpbft: Leader (%x) Tidy candidateBlock successfully, blockID:%x, txSum:%d\n", leaderAddress, candidateBlock.BlockID, len(candidateBlock.Transactions))
	return result, signature, candidateBlock
}
