package Network

import (
	"Elev_control"
	"fmt"
)

var isMaster bool

func MH_HandleOutgoingMsg(msgToNetwork, sendOrderCh chan UdpMessage, localStatusCh, updateElevsCh chan Elev_control.Elevator, sendBtnCallsCh, receiveBtnCallsCh chan [2]int) {
	var elev Elev_control.Elevator
	var msg UdpMessage
	for {
		if isMaster {
			select {
			case elev = <-localStatusCh:
				updateElevsCh <- elev
			case msg = <-sendOrderCh:
				msgToNetwork <- msg
			case btn_calls := <-sendBtnCallsCh:
				receiveBtnCallsCh <- btn_calls
			}
		} else {
			select {
			case elev = <-localStatusCh:
				msg.Data = elev
				msgToNetwork <- msg
			case new_call := <-sendBtnCallsCh:
				msg.Order_ID = 1
				msg.Order = new_call
				msgToNetwork <- msg
			}
		}
	}
}

func MH_HandleIncomingMsg(msgFromNetwork chan UdpMessage, updateElevsCh chan Elev_control.Elevator, receiveBtnCallsCh chan [2]int) {
	//var elev Elev_control.Elevator
	var msg UdpMessage
	for {
		msg = <-msgFromNetwork //msg.Data er elev
		//elev = msg.Data
		switch msg.Order_ID {
		case 0:
			updateElevsCh <- msg.Data
			//Elev_control.PrintElev(elev)
			fmt.Println("")
			break
		case 1:
			fmt.Println("Mottatt btn call fra slave")
			receiveBtnCallsCh <- msg.Order
		default:
			fmt.Println("Ordre mottas til: ", msg.Order_ID)
			Elev_control.Fsm_addOrder(msg.Order, msg.Order_ID)
		}
		//fmt.Println("Mottar udp melding")
	}
}

func MH_UpdateMasterStatus(isMasterFrom_Master_Slave bool) {
	isMaster = isMasterFrom_Master_Slave
}

func MH_send_new_order(to_elev int64, order [2]int, sendOrderCh chan UdpMessage) {
	var msg UdpMessage
	msg.Order_ID = to_elev
	msg.Order = order
	Elev_control.Fsm_addOrder(order, to_elev)
	sendOrderCh <- msg
}
