package gosdncon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	ofp13 "github.com/sangyun-han/monitor4sdn/controller/openflow/openflow13"
	"time"
)

// OFSwitch
type OFSwitch struct {
	buffer     chan *bytes.Buffer
	conn       *net.TCPConn
	dpid       uint64
	sendBuffer chan *ofp13.OFMessage
	ofVersion  string
	ports      int
	app        AppInterface
}

// Constructor
func NewSwitch(conn *net.TCPConn) *OFSwitch {
	fmt.Println("[OFSwitch] NewOFSwitch")

	sw := new(OFSwitch)
	sw.sendBuffer = make(chan *ofp13.OFMessage, 10)
	sw.conn = conn

	return sw
}

func (sw *OFSwitch) switchConnected() {
	sw.app.SwitchConnected(sw)

}

func (sw *OFSwitch) monitorLoop(interval int) {
	fmt.Println("[OFSwitch] startMonitoring")

	ticker := time.NewTicker(time.Second * time.Duration(interval))
	defer ticker.Stop()
	for t := range ticker.C {
		_ = t
		portStatsReq := ofp13.NewOfpPortStatsRequest(ofp13.OFPP_ANY, 0)
		sw.Send(portStatsReq)

		// TODO : will be added AggregateStatsRequest
		// OFPTT_ALL, OFPP_ANY, OFPG_ANY, mask=0
		aggregateStatsReq := ofp13.NewOfpAggregateStatsRequest(
			0,
			ofp13.OFPTT_ALL,
			ofp13.OFPP_ANY,
			ofp13.OFPG_ANY,
			0,
			0,
			ofp13.NewOfpMatch())
		sw.Send(aggregateStatsReq)

		// TODO : will be added FlowStatsRequest
		// OFPTT_ALL, OFPP_ANY, OFPG_ANY, mask=0
		flowStatsReq := ofp13.NewOfpFlowStatsRequest(
			0,
			ofp13.OFPTT_ALL,
			ofp13.OFPP_ANY,
			ofp13.OFPG_ANY,
			0,
			0,
			ofp13.NewOfpMatch())
		sw.Send(flowStatsReq)
	}
}

func (sw *OFSwitch) receiveLoop() {
	fmt.Println("[OFSwitch] receiveLoop")
	buf := make([]byte, 1024*64)
	for {
		size, err := sw.conn.Read(buf)
		if err != nil {
			fmt.Println("failed to read conn")
			fmt.Println(err)
			return
		}

		for i := 0; i < size; {
			msgLen := binary.BigEndian.Uint16(buf[i+2:])
			sw.handlePacket(buf[i : i+(int)(msgLen)])
			i += (int)(msgLen)
		}
	}
}

func (sw *OFSwitch) handlePacket(buf []byte) {
	// parse data
	msg := ofp13.Parse(buf[0:])

	if _, ok := msg.(*ofp13.OfpHello); ok {
		fmt.Println("[OFSwitch] handlePacket : First FeatureReq")
		// connected switch, session on
		featureReq := ofp13.NewOfpFeaturesRequest()
		sw.Send(featureReq)
	} else {
		// dispatch handler
		sw.dispatchHandler(msg)
	}
}

func (sw *OFSwitch) dispatchHandler(msg ofp13.OFMessage) {
	apps := GetAppManager().GetApplications()
	for _, app := range apps {
		switch msgi := msg.(type) {
		// if message is OfpHeader
		case *ofp13.OfpHeader:
			switch msgi.Type {
			// handle echo request
			case ofp13.OFPT_ECHO_REQUEST:
				if obj, ok := app.(Of13EchoRequestHandler); ok {
					obj.HandleEchoRequest(msgi, sw)
				}

			// handle echo reply
			case ofp13.OFPT_ECHO_REPLY:
				if obj, ok := app.(Of13EchoReplyHandler); ok {
					obj.HandleEchoReply(msgi, sw)
				}

			// handle Barrier reply
			case ofp13.OFPT_BARRIER_REPLY:
				if obj, ok := app.(Of13BarrierReplyHandler); ok {
					obj.HandleBarrierReply(msgi, sw)
				}
			default:
			}

		// Recv Error
		case *ofp13.OfpErrorMsg:
			if obj, ok := app.(Of13ErrorMsgHandler); ok {
				obj.HandleErrorMsg(msgi, sw)
			}

		// Recv RoleReply
		case *ofp13.OfpRole:
			if obj, ok := app.(Of13RoleReplyHandler); ok {
				obj.HandleRoleReply(msgi, sw)
			}

		// Recv GetAsyncReply
		case *ofp13.OfpAsyncConfig:
			if obj, ok := app.(Of13AsyncConfigHandler); ok {
				obj.HandleAsyncConfig(msgi, sw)
			}

		// case SwitchFeatures
		case *ofp13.OfpSwitchFeatures:
			if obj, ok := app.(Of13SwitchFeaturesHandler); ok {
				sw.dpid = msgi.DatapathId
				obj.HandleSwitchFeatures(msgi, sw)
			}

		// case GetConfigReply
		case *ofp13.OfpSwitchConfig:
			if obj, ok := app.(Of13SwitchConfigHandler); ok {
				obj.HandleSwitchConfig(msgi, sw)
			}
		// case PacketIn
		case *ofp13.OfpPacketIn:
			if obj, ok := app.(Of13PacketInHandler); ok {
				obj.HandlePacketIn(msgi, sw)
			}

		// case FlowRemoved
		case *ofp13.OfpFlowRemoved:
			if obj, ok := app.(Of13FlowRemovedHandler); ok {
				obj.HandleFlowRemoved(msgi, sw)
			}

		// case MultipartReply
		case *ofp13.OfpMultipartReply:
			switch msgi.Type {
			case ofp13.OFPMP_DESC:
				if obj, ok := app.(Of13DescStatsReplyHandler); ok {
					obj.HandleDescStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_FLOW:
				if obj, ok := app.(Of13FlowStatsReplyHandler); ok {
					obj.HandleFlowStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_AGGREGATE:
				if obj, ok := app.(Of13AggregateStatsReplyHandler); ok {
					obj.HandleAggregateStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_TABLE:
				if obj, ok := app.(Of13TableStatsReplyHandler); ok {
					obj.HandleTableStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_PORT_STATS:
				if obj, ok := app.(Of13PortStatsReplyHandler); ok {
					obj.HandlePortStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_QUEUE:
				if obj, ok := app.(Of13QueueStatsReplyHandler); ok {
					obj.HandleQueueStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_GROUP:
				if obj, ok := app.(Of13GroupStatsReplyHandler); ok {
					obj.HandleGroupStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_GROUP_DESC:
				if obj, ok := app.(Of13GroupDescStatsReplyHandler); ok {
					obj.HandleGroupDescStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_GROUP_FEATURES:
				if obj, ok := app.(Of13GroupFeaturesStatsReplyHandler); ok {
					obj.HandleGroupFeaturesStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_METER:
				if obj, ok := app.(Of13MeterStatsReplyHandler); ok {
					obj.HandleMeterStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_METER_CONFIG:
				if obj, ok := app.(Of13MeterConfigStatsReplyHandler); ok {
					obj.HandleMeterConfigStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_METER_FEATURES:
				if obj, ok := app.(Of13MeterFeaturesStatsReplyHandler); ok {
					obj.HandleMeterFeaturesStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_TABLE_FEATURES:
				if obj, ok := app.(Of13TableFeaturesStatsReplyHandler); ok {
					obj.HandleTableFeaturesStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_PORT_DESC:
				if obj, ok := app.(Of13PortDescStatsReplyHandler); ok {
					obj.HandlePortDescStatsReply(msgi, sw)
				}
			case ofp13.OFPMP_EXPERIMENTER:
				// TODO: implement
			default:
			}

		default:
			fmt.Println("UnSupport Message")
		}
	}
}

/**
 *
 */
func (sw *OFSwitch) Send(message ofp13.OFMessage) bool {
	fmt.Println("[OFSwitch] Send")
	// push data
	//(sw.sendBuffer) <- &message

	byteData := message.Serialize()
	_, err := sw.conn.Write(byteData)
	if err != nil {
		fmt.Println("failed to write conn : ", err)
		return false
	}

	return true
}
