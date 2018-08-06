package gosdncon

import (
	"fmt"
	"net"
	ofp13 "github.com/sangyun-han/monitor4sdn/controller/openflow/openflow13"
	"log"
	"os"
	"sync"
	"github.com/influxdata/influxdb/client/v2"
	"time"
	"encoding/json"
	"strconv"
	"io/ioutil"
)

type Configuration struct {
	TsdbName        string `json:"TsdbName"`
	TsdbAddr        string `json:"TsdbAddr"`
	Username         string `json:"Username"`
	Password         string `json:"Password"`
	MonitorInterval int `json:"MonitorInterval"`
}

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
	dbClient client.Client
	batchPoint client.BatchPoints
}

func NewOFController() *OFController {
		logger = log.New(os.Stdout, "[INFO][CONTROLLER] ", log.LstdFlags)
		logger.Println("[NewOFController]")
		ofc := new(OFController)
		ofc.echoInterval = 60
		ofc.switchDB = make(map[uint64]*OFSwitch)

	file, err := ioutil.ReadFile("conf.json")
	if err != nil {
		logger.Println("File error : ", err)
	}
	var config Configuration
	json.Unmarshal(file, &config)
	if err != nil {
		logger.Fatalln("Error : ", err)
	}

	if err != nil {
		logger.Fatalln("Error : ", err)
	}

	// Create a new client
	influxClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:config.TsdbAddr,
		Username:config.Username,
		Password:config.Password,
	})

	if err != nil {
		logger.Fatal(err)
	}
	//defer c.Close()
	ofc.dbClient = influxClient
	bpConfig := client.BatchPointsConfig{
		Database:config.TsdbName,
		Precision: "ns",
	}
	ofc.batchPoint, err = client.NewBatchPoints(bpConfig)
	if err != nil {
		logger.Fatal(err)
	}

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
	go sw.startMonitoring(1)

	// TODO add monitoring loop
}

func (c *OFController) writePoints(clnt client.Client, bp client.BatchPoints) {

}

func (c *OFController) HandleSwitchFeatures(msg *ofp13.OfpSwitchFeatures, sw *OFSwitch) {
	logger.Println("[HandleSwitchFeatures] DPID : ", sw.dpid)
	c.switchDB[sw.dpid] = sw
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
	logger.Println("[HandlePortStatsReply][",sw.dpid, "]")
	for _, mp := range msg.Body {
		if obj, ok := mp.(*ofp13.OfpPortStats); ok {
			tags := map[string]string{
				"dpid": strconv.FormatUint(sw.dpid, 10),
				"portNo": strconv.FormatUint(uint64(obj.PortNo), 10),
			}
			fields := map[string]interface{} {
				"RxPackets": int(obj.RxPackets),
				"TxPackets": int(obj.TxPackets),
				"RxBytes": int(obj.RxBytes),
				"TxBytes": int(obj.TxBytes),
				"RxDropped": int(obj.RxDropped),
				"TxDropped": int(obj.TxDropped),
				"RxErrors": int(obj.RxErrors),
				"TxErrors": int(obj.TxErrors),
				"RxFrameErr": int(obj.RxFrameErr),
				"RxOverErr": int(obj.RxOverErr),
				"RxCrcErr": int(obj.RxCrcErr),
				"Collisions": int(obj.Collisions),
				"DurationSec": int(obj.DurationSec),
				"DurationNSec": int(obj.DurationNSec),
			}
			point, err := client.NewPoint(
				"port_stats",
				tags,
				fields,
				time.Now(),
			)

			if err != nil {
				logger.Fatal("[HandlePortStatsReply][",sw.dpid, "] NewPoint Error : ", err)
			}
			c.batchPoint.AddPoint(point)

			if err := c.dbClient.Write(c.batchPoint); err != nil {
				logger.Fatal("[HandlePortStatsReply][",sw.dpid, "] AddPoint Error : ", err)
			}


		}
	}
}

func (c *OFController) HandleAggregateStatsReply(msg *ofp13.OfpMultipartReply, sw *OFSwitch) {
	logger.Println("[HandleAggregateStatsReply][",sw.dpid, "]")
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