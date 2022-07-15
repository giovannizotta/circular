package node

import (
	"circular/graph"
	"circular/util"
	"github.com/dgraph-io/badger/v3"
	"github.com/elementsproject/glightning/glightning"
	"log"
	"strconv"
	"time"
)

func (s *Node) SendPay(route *graph.Route, paymentHash string) (*glightning.SendPayFields, error) {
	defer util.TimeTrack(time.Now(), "node.SendPay")
	_, err := s.lightning.SendPayLite(route.ToLightningRoute(), paymentHash)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	result, err := s.lightning.WaitSendPay(paymentHash, PAYMENT_TIMEOUT)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return result, nil
}

func (s *Node) deleteIfOurs(paymentHash string) error {
	_, err := s.DB.Get(paymentHash)
	if err == badger.ErrKeyNotFound {
		return err // this payment was not made by us
	}
	err = s.DB.Delete(paymentHash)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *Node) OnPaymentFailure(sf *glightning.SendPayFailure) {
	if err := s.deleteIfOurs(sf.Data.PaymentHash); err != nil {
		return
	}
	direction := strconv.Itoa(sf.Data.ErringDirection)
	oppositeDirection := strconv.Itoa(sf.Data.ErringDirection ^ 0x1)
	channelId := sf.Data.ErringChannel + "/" + direction
	oppositeChannelId := sf.Data.ErringChannel + "/" + oppositeDirection
	log.Println("failed from " + s.Graph.Channels[oppositeChannelId].Source + " to " + s.Graph.Channels[oppositeChannelId].Destination)
	log.Printf("channel %s failed, opposite channel is %s\n", oppositeChannelId, channelId)
	log.Printf("code: %d, failcode: %d, failcodename: %s\n", sf.Code, sf.Data.FailCode, sf.Data.FailCodeName)
	// TODO: handle failure codes separately: right now we treat every failure as a liquidity failure, but it might not be the case
	s.Graph.Channels[channelId].Liquidity = sf.Data.MilliSatoshi - 1000000
	s.Graph.Channels[oppositeChannelId].Liquidity =
		s.Graph.Channels[oppositeChannelId].Satoshis*1000 - s.Graph.Channels[channelId].Liquidity
}

func (s *Node) OnPaymentSuccess(ss *glightning.SendPaySuccess) {
	s.deleteIfOurs(ss.PaymentHash)
}
