package main

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/client"
	walletseed "github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/mana"
	"github.com/iotaledger/hive.go/identity"
	"github.com/mr-tron/base58"
)

func main() {
	// Sample
	// dlaksdl
	// http://goshimmer.docs.iota.org/tutorials/wallet.html
	//
	goshimAPI := client.NewGoShimmerAPI("http://82.165.69.143:8080")
	myNode, _ := goshimAPI.Info()

	const seedAddr1 string = "59NYmXzp39JDnBgGcRDxj5fmKjpLx1TA1W5trJWRdtjV"
	const seedAddr2 string = "DkxqNM1r1crSFKhnFvarQa3Te2jijjh29voUdYocW8qU"

	// Request Faucet | 1 = Addr1 | 2 = Addr2 | 3 = Addr1 & 2
	faucetAddr := 4

	// Faucet

	// STEP 1 Seed //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	//var timestamp time.Time
	var accessPledgeID identity.ID
	var consensusPledgeID identity.ID
	var inputs ledgerstate.Inputs
	//var outputs ledgerstate.Outputs
	var version ledgerstate.TransactionEssenceVersion
	var timestamp time.Time

	// Version und Zeitstempel
	version = 0
	timestamp = time.Now()

	// Address 1
	seedBytesAddr1, _ := base58.Decode(seedAddr1) // ignoring error
	mySeed1 := walletseed.NewSeed(seedBytesAddr1)
	myAddr1 := mySeed1.Address(0)
	fmt.Println("My Address1: ", myAddr1.String())

	// Address 2
	seedBytesAddr2, _ := base58.Decode(seedAddr2)
	mySeed2 := walletseed.NewSeed(seedBytesAddr2)
	myAddr2 := mySeed2.Address(0)
	fmt.Println("My Address2: ", myAddr2.String())

	fmt.Println(goshimAPI.GetAddressUnspentOutputs(myAddr2.String()))
	// Convert NodeID for accessMana /////////////////////////////////////////////////////////////////////////

	pledgeID, err := mana.IDFromStr(myNode.IdentityID)
	if err != nil {
		fmt.Println("Error pledgeID")
		return
	}
	accessPledgeID = pledgeID
	consensusPledgeID = pledgeID

	fmt.Println("AccessPledgeID:", accessPledgeID, "\nConsensusPledgeID:", consensusPledgeID)
	// ///////////////////////////////////////////////////////////////////////////////////////////////////////

	switch {
	case faucetAddr == 1:
		messageID, err := goshimAPI.SendFaucetRequest(myAddr1.Base58(), 22, accessPledgeID.String(), consensusPledgeID.String())
		fmt.Println("Request Faucet from Address1")
		fmt.Println(messageID, err)
	case faucetAddr == 2:
		fmt.Println("Request Faucet from Address2")
		messageID, err := goshimAPI.SendFaucetRequest(myAddr2.Base58(), 22, accessPledgeID.String(), consensusPledgeID.String())
		fmt.Println(messageID, err)
	case faucetAddr == 3:
		fmt.Println("Request Faucet from Address1 & Address2")
		messageID1, err1 := goshimAPI.SendFaucetRequest(myAddr1.Base58(), 22, accessPledgeID.String(), consensusPledgeID.String())
		messageID2, err2 := goshimAPI.SendFaucetRequest(myAddr2.Base58(), 22, accessPledgeID.String(), consensusPledgeID.String())
		fmt.Println(messageID1, err1)
		fmt.Println(messageID2, err2)
	default:
		fmt.Println("No Faucet request")
	}

	// Prüft ob Guthaben zur verfügung steht und nicht bestätigt ist
	resp, _ := goshimAPI.PostAddressUnspentOutputs([]string{myAddr1.Base58()}) // ignoring error
	for _, output := range resp.UnspentOutputs[0].Outputs {
		fmt.Println("outputID:", output.Output.OutputID.Base58, "confirmed:", output.InclusionState.Confirmed)
	}

	// STEP 2 Transaction essence ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	//fmt.Println("Version:", version, "Zeitstempel:", timestamp)

	// Step 3 Inputs ///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Bereitstellen von nicht ausgegebenen Ausgaben

	resp2, _ := goshimAPI.GetAddressUnspentOutputs(myAddr1.Base58()) // ignoring error

	// iterate over unspent outputs of an address
	var out ledgerstate.Output
	for _, output := range resp2.Outputs {
		out, _ = output.ToLedgerstateOutput() // ignoring error
		balance, colorExist := out.Balances().Get(ledgerstate.ColorIOTA)
		fmt.Println(balance, colorExist)
	}

	out.Balances().ForEach(func(color ledgerstate.Color, balance uint64) bool {
		fmt.Println("Color:", color.Base58())
		fmt.Println("Balance:", balance)
		return true
	})

	inputs = ledgerstate.NewInputs(ledgerstate.NewUTXOInput(out.ID()))

	balance := ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: uint64(1000000),
	})

	output := ledgerstate.NewOutputs(ledgerstate.NewSigLockedColoredOutput(balance, myAddr2.Address()))
	kp := *mySeed1.KeyPair(0)
	txEssence := ledgerstate.NewTransactionEssence(version, timestamp, accessPledgeID, consensusPledgeID, inputs, output)
	signature := ledgerstate.NewED25519Signature(kp.PublicKey, kp.PrivateKey.Sign(txEssence.Bytes()))
	unlockBlock := ledgerstate.NewSignatureUnlockBlock(signature)
	tx := ledgerstate.NewTransaction(txEssence, ledgerstate.UnlockBlocks{unlockBlock})

	resp5, err := goshimAPI.PostTransaction(tx.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Transaction issued, txID:", resp5.TransactionID)

}
