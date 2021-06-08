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
	const seed string = "36af8Dovn8ctmt6WS6xWmtu4KP7GQxb1ZdXSAbbSRLS5"

	// Faucet

	// STEP 1 Seed //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Seed from ./cli-wallet
	seedBytes, _ := base58.Decode(seed) // ignoring error
	mySeed := walletseed.NewSeed(seedBytes)

	// Generate new address with index
	myAddr := mySeed.Address(0)

	messageID, err := goshimAPI.SendFaucetRequest(myAddr.Base58(), 22, "HwXLhewz61mK3QWiEdRhPt4kDLfmow7knyJrTqLw5rxz", "HwXLhewz61mK3QWiEdRhPt4kDLfmow7knyJrTqLw5rxz")
	fmt.Println(messageID, err)

	// My DevNet Seed
	// HnMtW6DsaPGFb4X11VzM9TFVSsJTmhBSPnxYhfm5f89C

	seedBytes2, _ := base58.Decode("DkxqNM1r1crSFKhnFvarQa3Te2jijjh29voUdYocW8qU")
	devNetByte := walletseed.NewSeed(seedBytes2)
	devNetAdresse := devNetByte.Address(0)

	fmt.Println("My Address: ", myAddr.String())
	fmt.Println("Zieladresse: ", devNetAdresse.String())

	// Prüft ob Guthaben zur verfügung steht und nicht bestätigt ist
	resp, _ := goshimAPI.PostAddressUnspentOutputs([]string{devNetAdresse.Base58()}) // ignoring error
	for _, output := range resp.UnspentOutputs[0].Outputs {
		fmt.Println("outputID:", output.Output.OutputID.Base58, "confirmed:", output.InclusionState.Confirmed)
	}

	// STEP 2 Transaction essence ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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

	//fmt.Println("Version:", version, "Zeitstempel:", timestamp)

	// Convert NodeID for accessMana /////////////////////////////////////////////////////////////////////////
	const nodeID string = "HwXLhewz61mK3QWiEdRhPt4kDLfmow7knyJrTqLw5rxz"
	pledgeID, err := mana.IDFromStr(nodeID)
	if err != nil {
		fmt.Println("Error pledgeID")
		return
	}
	accessPledgeID = pledgeID
	consensusPledgeID = pledgeID
	fmt.Println("AccessPledgeID:", accessPledgeID, "\nConsensusPledgeID:", consensusPledgeID)
	// ///////////////////////////////////////////////////////////////////////////////////////////////////////

	// Step 3 Inputs ///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Bereitstellen von nicht ausgegebenen Ausgaben

	resp2, _ := goshimAPI.GetAddressUnspentOutputs(devNetAdresse.Base58()) // ignoring error

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

	output := ledgerstate.NewOutputs(ledgerstate.NewSigLockedColoredOutput(balance, myAddr.Address()))
	kp := *devNetByte.KeyPair(0)
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

/*
func weg() {



}
*/
