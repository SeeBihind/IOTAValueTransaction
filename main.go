package main

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/client"
	walletseed "github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/mana"
	"github.com/mr-tron/base58"
)

func main() {
	// Sample
	// http://goshimmer.docs.iota.org/tutorials/wallet.html

	goshimAPI := client.NewGoShimmerAPI("http://82.165.69.143:8080")

	// Faucet
	// Seed from ./cli-wallet
	seedBytes, _ := base58.Decode("3heE6gdT3aAqRGZ2b7mARTjEQEvuVDACiC3CkDVQ1Nu6") // ignoring error
	mySeed := walletseed.NewSeed(seedBytes)
	// generate new address
	myAddr := mySeed.Address(0)

	a, err := goshimAPI.SendFaucetRequest(myAddr.Base58(), 22, "2GtxMQD94KvDH1SJPJV7icxofkyV1njuUZKtsqKmtux5", "2GtxMQD94KvDH1SJPJV7icxofkyV1njuUZKtsqKmtux5")

	fmt.Println("Faucet: ", a, err)
	fmt.Println("myAddr: ", myAddr)

	resp, _ := goshimAPI.PostAddressUnspentOutputs([]string{myAddr.Base58()}) // ignoring error
	for _, output := range resp.UnspentOutputs[0].Outputs {
		fmt.Println("outputID:", output.Output.OutputID.Base58, "confirmed:", output.InclusionState.Confirmed)
	}

	resp5, _ := goshimAPI.GetAddressUnspentOutputs(myAddr.Base58()) // ignoring error
	// iterate over unspent outputs of an address
	for _, output := range resp5.Outputs {
		var out ledgerstate.Output
		out, _ = output.ToLedgerstateOutput() // ignoring error

		out.Balances().ForEach(func(color ledgerstate.Color, balance uint64) bool {
			fmt.Println("Color:", color.Base58())
			fmt.Println("Balance:", balance)

			return true
		})

	}

	// build tx from previous step
	tx, err := buildTransaction()
	if err != nil {
		fmt.Println(err, "----")
		return
	}
	bytes := tx.Bytes()

	// print bytes
	fmt.Println("done")
	fmt.Println(string(bytes))

	resps, errs := goshimAPI.PostTransaction(bytes)
	if errs != nil {
		fmt.Println(errs)
		return
	}
	fmt.Println("Transaction issued, txID:", resps.TransactionID)

}

func buildTransaction() (tx *ledgerstate.Transaction, err error) {
	// node to pledge access mana.
	accessManaPledgeIDBase58 := "HwXLhewz61mK3QWiEdRhPt4kDLfmow7knyJrTqLw5rxz"
	accessManaPledgeID, err := mana.IDFromStr(accessManaPledgeIDBase58)
	if err != nil {
		return
	}

	// node to pledge consensus mana.
	consensusManaPledgeIDBase58 := "HwXLhewz61mK3QWiEdRhPt4kDLfmow7knyJrTqLw5rxz"
	consensusManaPledgeID, err := mana.IDFromStr(consensusManaPledgeIDBase58)
	if err != nil {
		return
	}

	/**
	  N.B to pledge mana to the node issuing the transaction, use empty pledgeIDs.
	  emptyID := identity.ID{}
	  accessManaPledgeID, consensusManaPledgeID := emptyID, emptyID
	  **/

	// destination address
	// 19nMrpMSZEqx3ntNXokxpsKmYrCW1yzPeHnR1kmkshpMG
	destAddressBase58 := "18GgPkjYRz9YqEQqUdBvbLQYjNRNybt8oYmyVXZxKcsQ9"
	destAddress, err := ledgerstate.AddressFromBase58EncodedString(destAddressBase58)
	if err != nil {
		fmt.Println(err, "destination address")
		return
	}

	// output to consume
	// 13oa4wXaURBJ2GewVYjrJxvyNifVw7XqehUFUHd5J9G6X
	outputIDBase58 := "32uvDAjEJDxT6YEaShQNAERs9Au5pMAkxFoCbuANBjtbjwV"
	out, err := ledgerstate.OutputIDFromBase58(outputIDBase58)
	if err != nil {
		fmt.Println(err, "output to consume")
		return
	}
	inputs := ledgerstate.NewInputs(ledgerstate.NewUTXOInput(out))

	// UTXO output.
	output := ledgerstate.NewSigLockedColoredOutput(ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: uint64(1337),
	}), destAddress)
	outputs := ledgerstate.NewOutputs(output)

	// build tx essence.
	txEssence := ledgerstate.NewTransactionEssence(0, time.Now(), accessManaPledgeID, consensusManaPledgeID, inputs, outputs)

	// sign.
	seed := walletseed.NewSeed([]byte("3heE6gdT3aAqRGZ2b7mARTjEQEvuVDACiC3CkDVQ1Nu6"))
	kp := seed.KeyPair(0)
	sig := ledgerstate.NewED25519Signature(kp.PublicKey, kp.PrivateKey.Sign(txEssence.Bytes()))
	unlockBlock := ledgerstate.NewSignatureUnlockBlock(sig)

	// build tx.
	tx = ledgerstate.NewTransaction(txEssence, ledgerstate.UnlockBlocks{unlockBlock})
	return
}
