package goexpress

//物流状态
const (
	//LogisticsStatusNone 未知
	LogisticsStatusNone = "暂无轨迹信息"
	//LogisticsStatusCollected 已揽件
	LogisticsStatusCollected = "已揽件"
	//LogisticsStatusShipping 运输在途
	LogisticsStatusShipping = "运输在途"
	//LogisticsStatusDelivered 已签收
	LogisticsStatusDelivered = "已签收"
	//LogisticsStatusException 问题件
	LogisticsStatusException = "问题件"
)
