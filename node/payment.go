package node

import (
	"circular/graph"
	"circular/util"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/elementsproject/glightning/glightning"
	"time"
)

const (
	FAILURE_PREFIX  = "f_"
	SUCCESS_PREFIX  = "s_"
	ROUTE_PREFIX    = "r_"
	TIMEOUT_PREFIX  = "timeout_"
	SENDPAY_TIMEOUT = 120 // 2 minutes
)

func (n *Node) SendPay(route *graph.Route, paymentHash string) (*glightning.SendPayFields, error) {
	defer util.TimeTrack(time.Now(), "node.SendPay", n.Logf)
	finalRoute := route.ToLightningRoute()

	n.Logln(glightning.Debug, "sending payment")
	if _, err := n.lightning.SendPayLite(finalRoute, paymentHash); err != nil {
		n.Logln(glightning.Unusual, err)
		return nil, err
	}

	n.Logln(glightning.Debug, "waiting for payment to be confirmed")
	result, err := n.lightning.WaitSendPay(paymentHash, SENDPAY_TIMEOUT)

	if err != nil {
		n.Logf(glightning.Debug, "%+v", err)
		n.Logln(glightning.Debug, "err.Error(): ", err.Error())

		// in case of timeout, there's some work to do
		if err.Error() == util.ErrSendPayTimeout.Error() {
			return n.manageTimeout(paymentHash)
		}

		// in case of WIRE_FEE_INSUFFICIENT, we return only if the last hop is the one who originated the error
		// in this way we make the rebalance fail if the last node changed fees, but treat
		// WIRE_FEE_INSUFFICIENT errors along the path as a liquidity failure
		if err.Error() == util.ErrWireFeeInsufficient.Error() {
			// we need to get the full error
			var paymentError = &glightning.PaymentError{}
			if errors.As(err, paymentError) {
				n.Logf(glightning.Debug, "WIRE_FEE_INSUFFICIENT error: %+v", paymentError)

				lastNode := finalRoute[len(finalRoute)-1].Id
				if lastNode == paymentError.Data.ErringNode {
					n.Logln(glightning.Debug, "last node is the node that caused the error")
					return nil, util.ErrWireFeeInsufficient
				}
			}
		}

		return nil, err
	}

	return result, nil
}

func (n *Node) manageTimeout(paymentHash string) (*glightning.SendPayFields, error) {
	// delete the preimage from the DB. In this way the payment will fail when the HTLC comes in
	n.Logln(glightning.Debug, "payment timed out, deleting preimage from database")
	if err := n.DB.Delete(paymentHash); err != nil {
		n.Logln(glightning.Unusual, err)
	}

	// save the failure in the DB. This will be used to update the liquidity
	n.Logln(glightning.Debug, "saving payment timeout to database")
	if err := n.DB.Set(TIMEOUT_PREFIX+paymentHash, []byte("timeout")); err != nil {
		n.Logln(glightning.Unusual, err)
	}

	return nil, util.ErrSendPayTimeout
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

func (n *Node) SaveToDb(key string, value any) error {
	if !n.saveStats {
		return nil
	}

	b, err := json.Marshal(value)
	if err != nil {
		n.Logln(glightning.Unusual, err)
		return err
	}

	err = n.DB.Set(key, b)
	if err != nil {
		n.Logln(glightning.Unusual, err)
		return err
	}

	return nil
}

func (n *Node) OnPaymentFailure(sf *glightning.SendPayFailure) {
	if err := n.deleteIfOurs(sf.Data.PaymentHash); err != nil {
		return // this payment was not made by us
	}

	// save to db
	if err := n.SaveToDb(FAILURE_PREFIX+sf.Data.PaymentHash, sf); err != nil {
		n.Logln(glightning.Unusual, err)
	}

	n.Logf(glightning.Debug, "code: %d, failcode: %d, failcodename: %s", sf.Code, sf.Data.FailCode, sf.Data.FailCodeName)

	// TODO: handle failure codes separately: right now we treat every failure as a liquidity failure, but it might not be the case
	n.LiquidityUpdateChan <- &LiquidityUpdate{
		Amount:         sf.Data.MilliSatoshi - util.Min(sf.Data.MilliSatoshi, 1000000),
		ShortChannelID: sf.Data.ErringChannel,
		Direction:      sf.Data.ErringDirection,
	}
}

func (n *Node) OnPaymentSuccess(ss *glightning.SendPaySuccess) {
	if err := n.deleteIfOurs(ss.PaymentHash); err != nil {
		return // this payment was not made by us
	}

	// save to db
	if err := n.SaveToDb(SUCCESS_PREFIX+ss.PaymentHash, ss); err != nil {
		n.Logln(glightning.Unusual, err)
	}
}
