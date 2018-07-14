package main

import (
	"fmt"
	ofp13 "github.com/sangyun-han/monitor4sdn/controller/openflow/openflow13"
	controller "github.com/sangyun-han/monitor4sdn/controller"
)

type SampleController struct {
	// add any paramter used in controller.
}

func NewSampleController() *SampleController {
	ofc := new(SampleController)
	return ofc
}

func (c *SampleController) HandleSwitchFeatures(msg *ofp13.OfpSwitchFeatures, dp *controller.Datapath) {
	// create match
	fmt.Println("[main][HandlwSwitchFeatures]")
	ethdst, _ := ofp13.NewOxmEthDst("00:00:00:00:00:00")
	if ethdst == nil {
		fmt.Println(ethdst)
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

	for i:= 1; i < 4; i++ {
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
	fmt.Println("[main][HandleFlowStatsReply]")

	for _, mp := range msg.Body {
		if obj, ok := mp.(*ofp13.OfpFlowStats); ok {
			fmt.Println("[main][HandleFlowStatsReply] ByteCount : ", obj.ByteCount)
			fmt.Println("[main][HandleFlowStatsReply] Instructions : ", obj.Instructions)
			fmt.Println("[main][HandleFlowStatsReply] Priority : ", obj.Priority)
		}
	}
}

func (c *SampleController) HandlwRoleReqply(msg *ofp13.OfpRole, dp *controller.Datapath) {
	fmt.Println("[main][HandlwRoleReqply]")
	fmt.Println("[main][HandlwRoleReqply] : ", msg.Header, msg.Role, msg.GenerationId)
}

func (c *SampleController)  HandlePortStatusReply(msg *ofp13.OfpPortStatus, dp *controller.Datapath) {
	fmt.Println("[main][HandlePortStatusReply]")
	fmt.Println("[main][HandlePortStatusReply] : ", msg.Desc)
}

func (c *SampleController) HandlePortStatsReply(msg *ofp13.OfpMultipartReply, dp *controller.Datapath) {
	fmt.Println("[main][HandlePortStatsReply]")
	fmt.Println("[main][HandlePortStatsReply] DPID : ")
	for _, mp := range msg.Body {
		if obj, ok := mp.(*ofp13.OfpPortStats); ok {
			fmt.Println("[main][HandlePortStatsReply] PortNo : ", obj.PortNo)
			fmt.Println("[main][HandlePortStatsReply] RxBytes : ", obj.RxBytes)
			fmt.Println("[main][HandlePortStatsReply] TxBytes : ", obj.TxBytes)
		}
	}
}

func (c *SampleController) HandleAggregateStatsReply(msg *ofp13.OfpMultipartReply, dp *controller.Datapath) {
	fmt.Println("[main][HandleAggregateStatsReply]")
	fmt.Println("Handle AggregateStats")
	for _, mp := range msg.Body {
		if obj, ok := mp.(*ofp13.OfpAggregateStats); ok {
			fmt.Println("[main][HandleAggregateStatsReply] PacketCount : ", obj.PacketCount)
			fmt.Println("[main][HandleAggregateStatsReply] ByteCount : ", obj.ByteCount)
			fmt.Println("[main][HandleAggregateStatsReply] FlowCount : ", obj.FlowCount)
		}
	}
}

func main() {
	fmt.Println("Start controller")
	ofc := NewSampleController()
	controller.GetAppManager().RegisterApplication(ofc)

	// start server
	controller.ServerLoop(controller.DEFAULT_PORT)
}
