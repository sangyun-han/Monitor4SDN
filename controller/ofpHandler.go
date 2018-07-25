package gosdncon

import (
	ofp13 "github.com/sangyun-han/monitor4sdn/controller/openflow/openflow13"
)

/*****************************************************/
/* OfpErrorMsg                                       */
/*****************************************************/
type Of13ErrorMsgHandler interface {
	HandleErrorMsg(*ofp13.OfpErrorMsg, *OFSwitch)
}

/*****************************************************/
/* Echo Message                                      */
/*****************************************************/
type Of13EchoRequestHandler interface {
	HandleEchoRequest(*ofp13.OfpHeader, *OFSwitch)
}

type Of13EchoReplyHandler interface {
	HandleEchoReply(*ofp13.OfpHeader, *OFSwitch)
}

/*****************************************************/
/* BarrierReply Message                              */
/*****************************************************/
type Of13BarrierReplyHandler interface {
	HandleBarrierReply(*ofp13.OfpHeader, *OFSwitch)
}

/*****************************************************/
/* OfpSwitchFeatures                                 */
/*****************************************************/
type Of13SwitchFeaturesHandler interface {
	HandleSwitchFeatures(*ofp13.OfpSwitchFeatures, *OFSwitch)
}

/*****************************************************/
/* OfpSwitchConfig                                   */
/*****************************************************/
type Of13SwitchConfigHandler interface {
	HandleSwitchConfig(*ofp13.OfpSwitchConfig, *OFSwitch)
}

/*****************************************************/
/* OfpPacketIn                                       */
/*****************************************************/
type Of13PacketInHandler interface {
	HandlePacketIn(*ofp13.OfpPacketIn, *OFSwitch)
}

/*****************************************************/
/* OfpFlowRemoved                                    */
/*****************************************************/
type Of13FlowRemovedHandler interface {
	HandleFlowRemoved(*ofp13.OfpFlowRemoved, *OFSwitch)
}

/*****************************************************/
/* OfpDescStatsReply                                 */
/*****************************************************/
type Of13DescStatsReplyHandler interface {
	HandleDescStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* OfpFlowStatsReply                                 */
/*****************************************************/
type Of13FlowStatsReplyHandler interface {
	HandleFlowStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* OfpAggregateStatsReply                            */
/*****************************************************/
type Of13AggregateStatsReplyHandler interface {
	HandleAggregateStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* OfpTableStatsReply                                */
/*****************************************************/
type Of13TableStatsReplyHandler interface {
	HandleTableStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* OfpPortStatsReply                                 */
/*****************************************************/
type Of13PortStatsReplyHandler interface {
	HandlePortStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* OfpQueueStatsReply                                */
/*****************************************************/
type Of13QueueStatsReplyHandler interface {
	HandleQueueStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* OfpGroupStatsReply                                */
/*****************************************************/
type Of13GroupStatsReplyHandler interface {
	HandleGroupStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* OfpGroupDescStatsReply                            */
/*****************************************************/
type Of13GroupDescStatsReplyHandler interface {
	HandleGroupDescStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* OfpGroupFeaturesStatsReply                        */
/*****************************************************/
type Of13GroupFeaturesStatsReplyHandler interface {
	HandleGroupFeaturesStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* OfpMeterStatsReply                                */
/*****************************************************/
type Of13MeterStatsReplyHandler interface {
	HandleMeterStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* OfpMeterConfigStatsReply                          */
/*****************************************************/
type Of13MeterConfigStatsReplyHandler interface {
	HandleMeterConfigStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* OfpMeterFeaturesStatsReply                        */
/*****************************************************/
type Of13MeterFeaturesStatsReplyHandler interface {
	HandleMeterFeaturesStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* OfpTableFeaturesStatsReply                        */
/*****************************************************/
type Of13TableFeaturesStatsReplyHandler interface {
	HandleTableFeaturesStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* OfpPortDescStatsReply                             */
/*****************************************************/
type Of13PortDescStatsReplyHandler interface {
	HandlePortDescStatsReply(*ofp13.OfpMultipartReply, *OFSwitch)
}

/*****************************************************/
/* RoleReply Message                                 */
/*****************************************************/
type Of13RoleReplyHandler interface {
	HandleRoleReply(*ofp13.OfpRole, *OFSwitch)
}

/*****************************************************/
/* GetAsyncReply Message                             */
/*****************************************************/
type Of13AsyncConfigHandler interface {
	HandleAsyncConfig(*ofp13.OfpAsyncConfig, *OFSwitch)
}
