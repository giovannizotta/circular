package node

import (
	"circular/graph"
	"circular/util"
	"encoding/json"
	"github.com/dgraph-io/badger/v3"
	"github.com/elementsproject/glightning/glightning"
	"time"
)

const (
	FAILURE_PREFIX  = "f_"
	SUCCESS_PREFIX  = "s_"
	TIMEOUT_PREFIX  = "timeout_"
	SENDPAY_TIMEOUT = 120 // 2 minutes
)

func (n *Node) SendPay(route *graph.Route, paymentHash string) (*glightning.SendPayFields, error) {
	defer util.TimeTrack(time.Now(), "node.SendPay")
	n.Logln(glightning.Debug, "sending payment")
	_, err := n.lightning.SendPayLite(route.ToLightningRoute(), paymentHash)
	if err != nil {
		n.Logln(glightning.Unusual, err)
		return nil, err
	}

	n.Logln(glightning.Debug, "waiting for payment to be confirmed")
	result, err := n.lightning.WaitSendPay(paymentHash, SENDPAY_TIMEOUT)
	if err != nil {
		if err.Error() == util.ErrSendPayTimeout.Error() {
			// delete the preimage from the DB. In this way the payment will fail when the HTLC comes in
			n.Logln(glightning.Debug, "payment timed out, deleting preimage from database")
			err = n.DB.Delete(paymentHash)
			if err != nil {
				n.Logln(glightning.Unusual, err)
			}

			// save the failure in the DB. This will be used to update the liquidity
			n.Logln(glightning.Debug, "saving payment failure to database")
			err = n.DB.Set(TIMEOUT_PREFIX+paymentHash, []byte("timeout"))
			if err != nil {
				n.Logln(glightning.Unusual, err)
			}
			return nil, util.ErrSendPayTimeout
		}
		n.Logln(glightning.Info, err)
		return nil, err
	}
	return result, nil
}

func (n *Node) deleteIfOurs(paymentHash string) error {
	key := paymentHash
	_, err := n.DB.Get(key)
	// check if this payment was made by us
	if err == badger.ErrKeyNotFound {
		// check if the payment timed out
		key = TIMEOUT_PREFIX + paymentHash
		_, err = n.DB.Get(key)
		if err == badger.ErrKeyNotFound {
			return err // this payment was not made by us
		}
	}

	err = n.DB.Delete(key)
	if err != nil {
		n.Logln(glightning.Unusual, err)
		return err
	}
	return nil
}

func (n *Node) OnPaymentFailure(sf *glightning.SendPayFailure) {
	if err := n.deleteIfOurs(sf.Data.PaymentHash); err != nil {
		return
	}

	// save to db
	bytes, err := json.Marshal(sf)
	if err != nil {
		n.Logln(glightning.Unusual, err)
	}
	err = n.DB.Set(FAILURE_PREFIX+sf.Data.PaymentHash, bytes)
	if err != nil {
		n.Logln(glightning.Unusual, err)
	}

	n.Logf(glightning.Debug, "code: %d, failcode: %d, failcodename: %s\n", sf.Code, sf.Data.FailCode, sf.Data.FailCodeName)

	// TODO: handle failure codes separately: right now we treat every failure as a liquidity failure, but it might not be the case
	n.LiquidityUpdateChan <- &LiquidityUpdate{
		Amount:         sf.Data.MilliSatoshi - 1000000,
		ShortChannelID: sf.Data.ErringChannel,
		Direction:      sf.Data.ErringDirection,
	}
}

func (n *Node) OnPaymentSuccess(ss *glightning.SendPaySuccess) {
	err := n.deleteIfOurs(ss.PaymentHash)
	if err != nil {
		return // this payment was not made by us
	}

	// save to db
	bytes, err := json.Marshal(ss)
	if err != nil {
		n.Logln(glightning.Unusual, err)
	}
	err = n.DB.Set(SUCCESS_PREFIX+ss.PaymentHash, bytes)
	if err != nil {
		n.Logln(glightning.Unusual, err)
	}
}
