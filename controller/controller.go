package gosdncon

import (
	"fmt"
	"net"
	ofp13 "github.com/sangyun-han/monitor4sdn/controller/openflow/openflow13"
)

var DEFAULT_PORT = 6653

/**
 * basic controller
 */
type OFController struct {
	echoInterval int32 // echo interval
	datapathList []Datapath
}

func NewOFController() *OFController {
	ofc := new(OFController)
	ofc.echoInterval = 60
	return ofc
}

func addDatapathToController(ofc *OFController, dp *Datapath) *OFController {
	controller := ofc
	controller.datapathList = append(controller.datapathList, *dp)
	return controller
}

//func (c *OFController) HandleHello(msg *ofp13.OfpHello, dp *Datapath) {
//	fmt.Println("recv Hello")
//	// send feature request
//	featureReq := ofp13.NewOfpFeaturesRequest()
//	(*dp).Send(featureReq)
//}

func (c *OFController) HandleSwitchFeatures(msg *ofp13.OfpSwitchFeatures, dp *Datapath) {
	fmt.Println("[controller][HSF]recv SwitchFeatures")
	// handle FeatureReply
	dp.datapathId = msg.DatapathId
	c.datapathList = append(c.datapathList, *dp)
	fmt.Println("[controller][HSF] DPID : ", msg.DatapathId)
}

func (c *OFController) HandleEchoRequest(msg *ofp13.OfpHeader, dp *Datapath) {
	fmt.Println("[controller][HER]recv EchoReq")
	// send EchoReply
	echo := ofp13.NewOfpEchoReply()
	(*dp).Send(echo)
}

func (c *OFController) ConnectionUp() {
	fmt.Println("[controller][ConnectionUp]")
	// handle connection up
}

func (c *OFController) ConnectionDown() {
	fmt.Println("[controller][ConnectionDown]")
	// handle connection down
}

func (c *OFController) sendEchoLoop() {
	fmt.Println("[controller][sendEchoLoop]")
	// send echo request forever
}

func ServerLoop(listenPort int) {
	var port int

	if listenPort <= 0 || listenPort >= 65536 {
		fmt.Println("Invalid port was specified. listen port must be between 0 - 65535.")
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
	// send hello
	hello := ofp13.NewOfpHello()
	_, err := conn.Write(hello.Serialize())
	if err != nil {
		fmt.Println(err)
	}

	// create datapath
	dp := NewDatapath(conn)

	// launch goroutine
	go dp.recvLoop()
	go dp.sendLoop()
}
