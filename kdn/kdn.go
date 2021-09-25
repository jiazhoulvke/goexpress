//Package kdn 快递鸟
package kdn

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jiazhoulvke/goexpress"
)

//常用快递编码
const (
	//ExpressCodeShunFeng 顺丰速运
	ExpressCodeShunFeng = "SF"
	//ExpressCodeBaiShi 百世快递
	ExpressCodeBaiShi = "HTKY"
	//ExpressCodeZhongTong 中通快递
	ExpressCodeZhongTong = "ZTO"
	//ExpressCodeShenTong 申通快递
	ExpressCodeShenTong = "STO"
	//ExpressCodeYuanTong 圆通速递
	ExpressCodeYuanTong = "YTO"
	//ExpressCodeYunDa 韵达速递
	ExpressCodeYunDa = "YD"
	//ExpressCodeYouZheng 邮政快递包裹
	ExpressCodeYouZheng = "YZPY"
	//ExpressCodeEMS EMS
	ExpressCodeEMS = "EMS"
	//ExpressCodeTianTian 天天快递
	ExpressCodeTianTian = "HHTT"
	//ExpressCodeYouSu 优速快递
	ExpressCodeYouSu = "UC"
	//ExpressCodeDeBang 德邦快递
	ExpressCodeDeBang = "DBL"
	//ExpressCodeJingDong 京东快递
	ExpressCodeJingDong = "JD"
	//ExpressCodeZhaiJiSong 宅急送
	ExpressCodeZhaiJiSong = "ZJS"
)

//APIURLTraces 物流轨迹即时查询接口地址
var APIURLTraces = "https://api.kdniao.com/Ebusiness/EbusinessOrderHandle.aspx" //正式地址
// var APIURLTraces = "http://sandboxapi.kdniao.cc:8080/kdniaosandbox/gateway/exterfaceInvoke.json" //测试地址

var (
	defautKDN *KDN
	inited    bool
)

//Init 初始化
func Init(c Config) error {
	var err error
	defautKDN, err = New(c)
	if err != nil {
		return err
	}
	inited = true
	return nil
}

//Traces 获取运单物流轨迹
func Traces(shipperCode, logisticCode, orderCode string, customerName string) (TracesResponse, error) {
	if !inited {
		return TracesResponse{}, fmt.Errorf("配置尚未初始化")
	}
	return defautKDN.Traces(shipperCode, logisticCode, orderCode, customerName)
}

//KDN 快递鸟
type KDN struct {
	config Config
}

//Traces 获取运单物流轨迹
func (k *KDN) Traces(shipperCode, logisticCode, orderCode string, customerName string) (TracesResponse, error) {
	var tracesResponse TracesResponse
	if shipperCode == "" {
		return tracesResponse, fmt.Errorf("快递公司编码不能为空")
	}
	if logisticCode == "" {
		return tracesResponse, fmt.Errorf("运单号不能为空")
	}
	reqData := TracesRequestData{
		LogisticCode: logisticCode,
		OrderCode:    orderCode,
		ShipperCode:  shipperCode,
		CustomerName: customerName,
	}
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return tracesResponse, err
	}
	requestData := string(jsonData)
	sign, err := Sign(requestData, k.config.AppKey)
	if err != nil {
		return tracesResponse, err
	}
	values := url.Values{}
	values.Set("DataSign", sign)
	values.Set("DataType", "json")
	values.Set("EBusinessID", k.config.EBusinessID)
	values.Set("RequestData", url.PathEscape(requestData))
	values.Set("RequestType", "1002")
	resp, err := http.PostForm(APIURLTraces, values)
	if err != nil {
		return tracesResponse, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&tracesResponse); err != nil {
		return tracesResponse, err
	}
	if !tracesResponse.IsSuccess() {
		return tracesResponse, fmt.Errorf("%s", tracesResponse.Reason)
	}
	return tracesResponse, nil
}

//Config 配置
type Config struct {
	//EBusinessID 商户ID
	EBusinessID string
	//AppKey 密钥
	AppKey string
}

//TracesRequestData 查询物流轨迹请求数据
type TracesRequestData struct {
	OrderCode string `json:"OrderCode"`
	// OrderCode    string `json:"OrderCode,omitempty"`
	ShipperCode  string `json:"ShipperCode"`
	LogisticCode string `json:"LogisticCode"`
	CustomerName string `json:"CustomerName"`
}

//Trace 物流轨迹信息
type Trace struct {
	AcceptTime    string `json:"AcceptTime"`
	AcceptStation string `json:"AcceptStation"`
	Remark        string `json:"Remark"`
}

//TracesResponse 查询物流轨迹返回数据
type TracesResponse struct {
	EBusinessID  string `json:"EBusinessID"`
	OrderCode    string `json:"OrderCode"`
	ShipperCode  string `json:"ShipperCode"`
	LogisticCode string `json:"LogisticCode"`
	//Success 是否成功
	//快递鸟的菜鸟程序员在成功时返回bool类型的true，失败时返回字符串类型的"false",
	//所以只好用interface{}来接收这个值
	Success interface{} `json:"Success"`
	State   string      `json:"State"`
	Reason  string      `json:"Reason,omitempty"`
	Traces  []Trace     `json:"Traces"`
}

//IsSuccess 是否成功
func (t TracesResponse) IsSuccess() bool {
	switch success := t.Success.(type) {
	case bool:
		if !success {
			return false
		}
	case string:
		if success != "true" {
			return false
		}
	default:
		return false
	}
	return true
}

//LogisticsStatus 运单状态
func (t TracesResponse) LogisticsStatus() string {
	switch t.State {
	case "1":
		return goexpress.LogisticsStatusCollected
	case "2":
		return goexpress.LogisticsStatusShipping
	case "3":
		return goexpress.LogisticsStatusDelivered
	case "4":
		return goexpress.LogisticsStatusException
	default:
		return goexpress.LogisticsStatusNone
	}
}

//New 返回一个实例
func New(c Config) (*KDN, error) {
	if c.EBusinessID == "" {
		return nil, fmt.Errorf("商户ID[EBusinessID]不能为空")
	}
	if c.AppKey == "" {
		return nil, fmt.Errorf("AppKey不能为空")
	}
	return &KDN{
		config: c,
	}, nil
}

//Sign 签名
func Sign(reqData string, appKey string) (string, error) {
	h := md5.New()
	_, err := h.Write([]byte(reqData + appKey))
	if err != nil {
		return "", err
	}
	b64 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", h.Sum(nil))))
	sign := url.PathEscape(b64)
	return sign, nil
}
