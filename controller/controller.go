package gosdncon

import (
	"fmt"
	"net"
	ofp13 "github.com/sangyun-han/monitor4sdn/controller/openflow/openflow13"
	"log"
	"os"
)

var DEFAULT_PORT = 6653
var logger *log.Logger

/**
 * basic controller
 */
type OFController struct {
	echoInterval int32 // echo interval
	datapathList []OFSwitch
}

func NewOFController() *OFController {
	logger = log.New(os.Stdout, "[INFO][CONTROLLER] ", log.LstdFlags)
	logger.Println("[NewOFController]")
	ofc := new(OFController)
	ofc.echoInterval = 60
	return ofc
}

func (c *OFController) ConnectionUp() {
	logger.Println("[controller][ConnectionUp]")
	// handle connection up
}

func (c *OFController) ConnectionDown() {
	logger.Println("[controller][ConnectionDown]")
	// handle connection down
}

func (c *OFController) sendEchoLoop() {
	logger.Println("[controller][sendEchoLoop]")
	// send echo request forever
}

func ServerLoop(listenPort int) {
	//logger = log.New(os.Stdout, "[INFO][CONTROLLER] ", log.LstdFlags)
	logger.Println("[ServerLoop]")
	var port int

	if listenPort <= 0 || listenPort >= 65536 {
		logger.Println("Invalid port was specified. listen port must be between 0 - 65535.")
		return
	}
	port = listenPort

	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", port))
	listener, err := net.ListenTCP("tcp", tcpAddr)

	ofc := NewOFController()
	GetAppManager().RegisterApplication(ofc)

	if err != nil {
		return
	}

	// wait for connect from switch
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			return
		}
		go handleConnection(conn)
	}
}

/**
 *
 */
func handleConnection(conn *net.TCPConn) {
	logger.Println("[handleConnection]")
	// send hello
	hello := ofp13.NewOfpHello()
	_, err := conn.Write(hello.Serialize())
	if err != nil {
		logger.Println(err)
	}

	// create datapath
	sw := NewSwitch(conn)

	// launch goroutine
	go sw.receiveLoop()
	go sw.sendLoop()
}


func (c *OFController) HandleSwitchFeatures(msg *ofp13.OfpSwitchFeatures, sw *OFSwitch) {
	logger.Println("[HandleSwitchFeatures] DPID : ", msg.DatapathId)
	// create match
	logger.Println("[HandlwSwitchFeatures]")
	ethdst, _ := ofp13.NewOxmEthDst("00:00:00:00:00:00")
	if ethdst == nil {
		logger.Println(ethdst)
		return
	}
	match := ofp13.NewOfpMatch()
	match.Append(ethdst)

	// create Instruction
	instruction := ofp13.NewOfpInstructionActions(ofp13.OFPIT_APPLY_ACTIONS)

	// create actions
	seteth, _ := ofp13.NewOxmEthDst("11:22:33:44:55:66")
	instruction.Append(ofp13.NewOfpActionSetField(seteth))

	// append Instruction
	instructions := make([]ofp13.OfpInstruction, 0)
	instructions = append(instructions, instruction)

	// create flow mod
	fm := ofp13.NewOfpFlowModModify(
		0, // cookie
		0, // cookie mask
		0, // tableid
		0, // priority
		ofp13.OFPFF_SEND_FLOW_REM,
		match,
		instructions,
	)

	// send FlowMod
	sw.Send(fm)

	// Create and send AggregateStatsRequest
	mf := ofp13.NewOfpMatch()
	mf.Append(ethdst)
	mp := ofp13.NewOfpAggregateStatsRequest(0, 0, ofp13.OFPP_ANY, ofp13.OFPG_ANY, 0, 0, mf)
	sw.Send(mp)

	for i:= 1; i < 10; i++ {
		mp2 := ofp13.NewOfpPortStatsRequest(uint32(i), 0)
		sw.Send(mp2)
	}

	//mp3 := ofp13.NewOfpFlowStatsRequest(0, 0, ofp13.OFPP_ANY, ofp13.OFPG_ANY, 0, 0, ofp13.NewOfpMatch())
	//sw.Send(mp3)

	//mp4 := ofp13.NewOfpFeaturesRequest()
	//sw.Send(mp4)

	//mp4 := ofp13.NewOfpPortStatus()
	//sw.Send(mp4)

	//mp3 := ofp13.NewOfpRoleRequest(uint32(0x00000003), uint64(0))
	//sw.Send(mp3)
}

func (c *OFController) HandleFlowStatsReply(msg *ofp13.OfpMultipartReply, sw *OFSwitch) {
	logger.Println("[HandleFlowStatsReply]")

	for _, mp := range msg.Body {
		if obj, ok := mp.(*ofp13.OfpFlowStats); ok {
			logger.Println("[HandleFlowStatsReply] ByteCount : ", obj.ByteCount)
			logger.Println("[HandleFlowStatsReply] Instructions : ", obj.Instructions)
			logger.Println("[HandleFlowStatsReply] Priority : ", obj.Priority)
		}
	}
}

func (c *OFController) HandlwRoleReqply(msg *ofp13.OfpRole, sw *OFSwitch) {
	logger.Println("[HandlwRoleReqply]")
	logger.Println("[HandlwRoleReqply] : ", msg.Header, msg.Role, msg.GenerationId)
}

func (c *OFController)  HandlePortStatusReply(msg *ofp13.OfpPortStatus, sw *OFSwitch) {
	logger.Println("[HandlePortStatusReply]")
	logger.Println("[HandlePortStatusReply] : ", msg.Desc)
}

func (c *OFController) HandlePortStatsReply(msg *ofp13.OfpMultipartReply, sw *OFSwitch) {
	logger.Println("[HandlePortStatsReply]")
	logger.Println("[HandlePortStatsReply] DPID : ")
	for _, mp := range msg.Body {
		if obj, ok := mp.(*ofp13.OfpPortStats); ok {
			logger.Println("[HandlePortStatsReply] PortNo : ", obj.PortNo)
			logger.Println("[HandlePortStatsReply] RxBytes : ", obj.RxBytes)
			logger.Println("[HandlePortStatsReply] TxBytes : ", obj.TxBytes)
		}
	}
}

func (c *OFController) HandleAggregateStatsReply(msg *ofp13.OfpMultipartReply, sw *OFSwitch) {
	logger.Println("[HandleAggregateStatsReply]")
	logger.Println("Handle AggregateStats")
	for _, mp := range msg.Body {
		if obj, ok := mp.(*ofp13.OfpAggregateStats); ok {
			logger.Println("[HandleAggregateStatsReply] PacketCount : ", obj.PacketCount)
			logger.Println("[HandleAggregateStatsReply] ByteCount : ", obj.ByteCount)
			logger.Println("[HandleAggregateStatsReply] FlowCount : ", obj.FlowCount)
		}
	}
}

//func (c *OFController) HandleHello(msg *ofp13.OfpHello, sw *OFSwitch) {
//	logger.Println("recv Hello")
//	// send feature request
//	featureReq := ofp13.NewOfpFeaturesRequest()
//	(*sw).Send(featureReq)
//}

//func (c *OFController) HandlePortStatsReply(msg *ofp13.OfpMultipartReply, sw *OFSwitch) {
//	logger.Println("[HandlePortStatsReply]")
//	logger.Println("[HandlePortStatsReply] DPID : ")
//	for _, mp := range msg.Body {
//		if obj, ok := mp.(*ofp13.OfpPortStats); ok {
//			logger.Println("[HandlePortStatsReply] PortNo : ", obj.PortNo)
//			logger.Println("[HandlePortStatsReply] RxBytes : ", obj.RxBytes)
//			logger.Println("[HandlePortStatsReply] TxBytes : ", obj.TxBytes)
//		}
//	}
//}

//func (c *OFController) HandleSwitchFeatures(msg *ofp13.OfpSwitchFeatures, sw *OFSwitch) {
//	logger.Println("[controller][HSF]recv SwitchFeatures")
//	// handle FeatureReply
//	sw.datapathId = msg.DatapathId
//	c.datapathList = append(c.datapathList, *sw)
//	logger.Println("[controller][HSF] DPID : ", msg.DatapathId)
//}

func (c *OFController) HandleEchoRequest(msg *ofp13.OfpHeader, sw *OFSwitch) {
	logger.Println("[controller][HER]recv EchoReq")
	// send EchoReply
	echo := ofp13.NewOfpEchoReply()
	(*sw).Send(echo)
}