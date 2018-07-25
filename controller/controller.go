package gosdncon

import (
	"fmt"
	"net"
	ofp13 "github.com/sangyun-han/monitor4sdn/controller/openflow/openflow13"
	"log"
	"os"
	"sync"
)

var DEFAULT_PORT = 6653
var logger *log.Logger

type AppInterface interface {
	// A switch connected to the SDN controller
	SwitchConnected(sw *OFSwitch)

	// Switch disconnected from the SDN controller
	SwitchDisconnected(sw *OFSwitch)
}

/**
 * basic controller
 */
type OFController struct {
	app          AppInterface
	echoInterval int32 // echo interval
	switchDB     map[uint64]*OFSwitch
	waitGroup    sync.WaitGroup
}

func NewOFController() *OFController {
	logger = log.New(os.Stdout, "[INFO][CONTROLLER] ", log.LstdFlags)
	logger.Println("[NewOFController]")
	ofc := new(OFController)
	ofc.echoInterval = 60
	ofc.switchDB = make(map[uint64]*OFSwitch)
	return ofc
}


func Listen(listenPort int) {
	logger = log.New(os.Stdout, "[INFO][CONTROLLER] ", log.LstdFlags)
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
		ofc.waitGroup.Add(1)
		go ofc.handleConnection(conn)
	}
}


func (c *OFController) handleConnection(conn *net.TCPConn) {
	logger.Println("[handleConnection]")

	defer c.waitGroup.Done()

	// Send openflow 1.3 Hello by default
	hello := ofp13.NewOfpHello()
	_, err := conn.Write(hello.Serialize())
	if err != nil {
		logger.Println(err)
	}

	// create ofSwitch
	sw := NewSwitch(conn)

	// launch goroutine
	go sw.receiveLoop()

	// TODO add monitoring loop
}


func (c *OFController) HandleSwitchFeatures(msg *ofp13.OfpSwitchFeatures, sw *OFSwitch) {
	logger.Println("[HandleSwitchFeatures] DPID : ", sw.dpid)
	c.switchDB[sw.dpid] = sw

	//for i:= 1; i < 10; i++ {
	//	mp2 := ofp13.NewOfpPortStatsRequest(uint32(i), 0)
	//	sw.Send(mp2)
	//}

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

func (c *OFController) HandleHello(msg *ofp13.OfpHello, sw *OFSwitch) {
	logger.Println("recv Hello")
	// send feature request
	featureReq := ofp13.NewOfpFeaturesRequest()
	(*sw).Send(featureReq)
}

func (c *OFController) HandleEchoRequest(msg *ofp13.OfpHeader, sw *OFSwitch) {
	logger.Println("[controller][HER]recv EchoReq")
	// send EchoReply
	echo := ofp13.NewOfpEchoReply()
	(*sw).Send(echo)
}