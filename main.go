package main

import (
	ofp13 "github.com/sangyun-han/monitor4sdn/controller/openflow/openflow13"
	controller "github.com/sangyun-han/monitor4sdn/controller"
	"log"
	"os"
)

var logger *log.Logger

type SampleController struct {
	// add any paramter used in controller.
}

func NewSampleController() *SampleController {
	ofc := new(SampleController)
	return ofc
}

func (c *SampleController) HandleSwitchFeatures(msg *ofp13.OfpSwitchFeatures, dp *controller.Datapath) {
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
	dp.Send(fm)

	// Create and send AggregateStatsRequest
	mf := ofp13.NewOfpMatch()
	mf.Append(ethdst)
	mp := ofp13.NewOfpAggregateStatsRequest(0, 0, ofp13.OFPP_ANY, ofp13.OFPG_ANY, 0, 0, mf)
	dp.Send(mp)

	for i:= 1; i < 10; i++ {
		mp2 := ofp13.NewOfpPortStatsRequest(uint32(i), 0)
		dp.Send(mp2)
	}

	//mp3 := ofp13.NewOfpFlowStatsRequest(0, 0, ofp13.OFPP_ANY, ofp13.OFPG_ANY, 0, 0, ofp13.NewOfpMatch())
	//dp.Send(mp3)

	//mp4 := ofp13.NewOfpFeaturesRequest()
	//dp.Send(mp4)

	//mp4 := ofp13.NewOfpPortStatus()
	//dp.Send(mp4)

	//mp3 := ofp13.NewOfpRoleRequest(uint32(0x00000003), uint64(0))
	//dp.Send(mp3)
}

func (c *SampleController) HandleFlowStatsReply(msg *ofp13.OfpMultipartReply, dp *controller.Datapath) {
	logger.Println("[HandleFlowStatsReply]")

	for _, mp := range msg.Body {
		if obj, ok := mp.(*ofp13.OfpFlowStats); ok {
			logger.Println("[HandleFlowStatsReply] ByteCount : ", obj.ByteCount)
			logger.Println("[HandleFlowStatsReply] Instructions : ", obj.Instructions)
			logger.Println("[HandleFlowStatsReply] Priority : ", obj.Priority)
		}
	}
}

func (c *SampleController) HandlwRoleReqply(msg *ofp13.OfpRole, dp *controller.Datapath) {
	logger.Println("[HandlwRoleReqply]")
	logger.Println("[HandlwRoleReqply] : ", msg.Header, msg.Role, msg.GenerationId)
}

func (c *SampleController)  HandlePortStatusReply(msg *ofp13.OfpPortStatus, dp *controller.Datapath) {
	logger.Println("[HandlePortStatusReply]")
	logger.Println("[HandlePortStatusReply] : ", msg.Desc)
}

func (c *SampleController) HandlePortStatsReply(msg *ofp13.OfpMultipartReply, dp *controller.Datapath) {
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

func (c *SampleController) HandleAggregateStatsReply(msg *ofp13.OfpMultipartReply, dp *controller.Datapath) {
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

func main() {
	logger = log.New(os.Stdout, "[INFO][MAIN] ", log.LstdFlags)
	logger.Println("Start controller")
	ofc := NewSampleController()
	controller.GetAppManager().RegisterApplication(ofc)
	
	// start server
	controller.ServerLoop(controller.DEFAULT_PORT)

}
