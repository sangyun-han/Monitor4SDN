package gosdncon

import (
	"fmt"
	"net"
	ofp13 "github.com/sangyun-han/monitor4sdn/controller/openflow/openflow13"
	"log"
	"os"
	"sync"

	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	TsdbName        string `json:"TsdbName"`
	TsdbAddr        string `json:"TsdbAddr"`
	Username         string `json:"Username"`
	Password         string `json:"Password"`
	MonitorInterval int `json:"MonitorInterval"`
}

/**
 * basic controller
 */
type OFController struct {
	app             AppInterface
	echoInterval    int32 // echo interval
	switchDB        map[uint64]*OFSwitch
	waitGroup       sync.WaitGroup
	monitorInterval int
}

var DEFAULT_PORT = 6653
var logger *log.Logger

type AppInterface interface {
	// A switch connected to the SDN controller
	SwitchConnected(sw *OFSwitch)

	// Switch disconnected from the SDN controller
	SwitchDisconnected(sw *OFSwitch)
}

func NewOFController() *OFController {
	logger = log.New(os.Stdout, "[INFO][CONTROLLER] ", log.LstdFlags)
	logger.Println("[NewOFController]")
	ofc := new(OFController)
	ofc.echoInterval = 60
	ofc.switchDB = make(map[uint64]*OFSwitch)

	config, err := loadConfig("conf.json")
	if err != nil {
		logger.Fatal("failed to load configuration file: ", err)
	}

	ofc.monitorInterval = config.MonitorInterval

	// Create a new client


	if err != nil {
		logger.Fatal("[NewOFController][InfluxClient] ", err)
	}
	//defer c.Close()


	return ofc
}

func loadConfig(filename string) (Configuration, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Println("File error : ", err)
		return Configuration{}, err
	}
	var config Configuration
	json.Unmarshal(file, &config)

	return config, nil
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
	go sw.monitorLoop(c.monitorInterval)
}

func (c *OFController) HandleSwitchFeatures(msg *ofp13.OfpSwitchFeatures, sw *OFSwitch) {
	logger.Println("[HandleSwitchFeatures] DPID : ", sw.dpid)
	c.switchDB[sw.dpid] = sw
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
	logger.Println("[HandlePortStatsReply][",sw.dpid, "]")
	for _, mp := range msg.Body {
		if obj, ok := mp.(*ofp13.OfpPortStats); ok {
			logger.Println("[HandlePortStatsReply][",sw.dpid, "] TxBytes", obj.TxBytes)
			logger.Println("[HandlePortStatsReply][",sw.dpid, "] TxBytes", obj.RxBytes)
		}
	}
}

func (c *OFController) HandleAggregateStatsReply(msg *ofp13.OfpMultipartReply, sw *OFSwitch) {
	logger.Println("[HandleAggregateStatsReply][",sw.dpid, "]")
	for _, mp := range msg.Body {
		if obj, ok := mp.(*ofp13.OfpAggregateStats); ok {
			logger.Println("[HandleAggregateStatsReply][", sw.dpid, "] PacketCount : ", obj.PacketCount)
			logger.Println("[HandleAggregateStatsReply][", sw.dpid, "] ByteCount : ", obj.ByteCount)
			logger.Println("[HandleAggregateStatsReply][", sw.dpid, "] FlowCount : ", obj.FlowCount)

		}
	}
}

func (c *OFController) HandleFlowStatsReply(msg *ofp13.OfpMultipartReply, sw *OFSwitch) {
	logger.Println("[HandleFlowStatsReply][",sw.dpid, "]")

	for _, mp := range msg.Body {
		if obj, ok := mp.(*ofp13.OfpFlowStats); ok {
			logger.Println("[HandleFlowStatsReply] ByteCount : ", obj.ByteCount)
			logger.Println("[HandleFlowStatsReply] Instructions : ", obj.Instructions)
			logger.Println("[HandleFlowStatsReply] Priority : ", obj.Priority)
		}
	}
}

func (c *OFController) HandleHelloMsg(msg *ofp13.OfpHello, sw *OFSwitch) {
	logger.Println("recv Hello")
	// send feature request
	featureReq := ofp13.NewOfpFeaturesRequest()
	(*sw).Send(featureReq)
}

func (c *OFController) HandleEchoRequest(msg *ofp13.OfpHeader, sw *OFSwitch) {
	logger.Println("[controller] recv EchoReq")
	// send EchoReply
	echo := ofp13.NewOfpEchoReply()
	(*sw).Send(echo)
}

func (c *OFController) HandleEchoReply(msg *ofp13.OfpHeader, sw *OFSwitch) {
	logger.Println("[controller] recv Echo Reply")
}