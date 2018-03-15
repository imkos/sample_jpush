package jpush

import (
	"errors"
	"jsonc"
)

type (
	//推送结构体
	JPushClient struct {
		appKey       string
		masterSecret string
		headers      map[string]string
		wrapper      *Playload
	}
	RateLimitInfo struct {
		RateLimitQuota     int
		RateLimitRemaining int
		RateLimitReset     int
	}
	ErrorResult struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	ResponseBase struct {
		// HTTP 状态码
		StatusCode int
		// 频率限制相关
		RateLimitInfo *RateLimitInfo
		// 错误相关
		Error *ErrorResult `json:"error"`
	}
	PushResult struct {
		ResponseBase
		// 成功时 msg_id 是 string 类型。。。
		// 失败时 msg_id 是 int 类型。。。
		MsgId  interface{} `json:"msg_id"`
		SendNo string      `json:"sendno"`
	}
	//推送playload结构体
	Playload struct {
		Platform     interface{}   `json:"platform"`
		Audience     interface{}   `json:"audience"`
		Notification *Notification `json:"notification,omitempty"`
		Message      *Message      `json:"message,omitempty"`
		SmsMessage   *SmsMessage   `json:"sms_message,omitempty"`
		Options      *Options      `json:"options,omitempty"`
	}
	Platform struct {
		isAll bool
		value []string
	}
	Audience struct {
		isAll bool // 是否推送给所有对象，如果是，value 无效
		value map[string][]string
	}
	Message struct {
		Content     string `json:"msg_content"`
		Title       string `json:"title,omitempty"`
		ContentType string `json:"content_type,omitempty"`
		//Extras@ json object: map[string]interface{}
		Extras interface{} `json:"extras,omitempty"`
	}
	platformNotification struct {
		Alert string `json:"alert"` // required
		//Extras@ json object: map[string]interface{}
		Extras interface{} `json:"extras,omitempty"`
	}
	// Android 平台上的通知。
	AndroidNotification struct {
		platformNotification
		Title     string `json:"title,omitempty"`
		BuilderId int    `json:"builder_id,omitempty"`
	}
	IosNotification struct {
		platformNotification
		Sound            string `json:"sound,omitempty"`
		Badge            int    `json:"badge,omitempty"`
		ContentAvailable bool   `json:"content-available,omitempty"`
		Category         string `json:"category,omitempty"`
	}
	WinphoneNotification struct {
		platformNotification
		Title    string `json:"title,omitempty"`
		OpenPage string `json:"_open_page,omitempty"`
	}
	Notification struct {
		Alert    string                `json:"alert,omitempty"`
		Android  *AndroidNotification  `json:"android,omitempty"`
		Ios      *IosNotification      `json:"ios,omitempty"`
		Winphone *WinphoneNotification `json:"winphone,omitempty"`
	}
	Options struct {
		SendNo          int   `json:"sendno,omitempty"`
		TimeToLive      int   `json:"time_to_live,omitempty"`
		OverrideMsgId   int64 `json:"override_msg_id,omitempty"`
		ApnsProduction  bool  `json:"apns_production"`
		BigPushDuration int   `json:"big_push_duration,omitempty"`
	}
	SmsMessage struct {
	}
	//设备接口结构体
	QueryDeviceResult struct {
		ResponseBase
		// 设备的所有属性，包含tags, alias
		Tags  []string `json:"tags"`
		Alias string   `json:"alias"`
	}
	tags struct {
		Add    []string `json:"add,omitempty"`
		Remove []string `json:"remove,omitempty"`
		Clear  bool     `json:"-"`
	}
	DeviceUpdate struct {
		// 支持 add, remove 或者空字符串。
		// 当 tags 参数为空字符串的时候，表示清空所有的 tags；
		// add/remove 下是增加或删除指定的 tags
		Tags tags
		// 更新设备的别名属性；当别名为空串时，删除指定设备的别名；
		Alias string
		// 手机号码
		Mobile string
	}
	deviceUpdateWrapper struct {
		Tags   interface{} `json:"tags"`
		Alias  string      `json:"alias"`
		Mobile string      `json:"mobile"`
	}
)

const (
	CHARSET             = "UTF-8"
	CONTENT_TYPE_JSON   = "application/json"
	USER_AGENT          = "sctek.com-go-jpush-api-client"
	CONNECTION_ALIVE    = "keep-alive"
	DEF_CONNECT_TIMEOUT = 10 //seconds
	//DEF_SOCKET_TIMEOUT  = 30
	//极光推送url
	PUSH_HOST = "https://api.jpush.cn"
	PUSH_URL  = PUSH_HOST + "/v3/push"
	//PUSH_VALIDATE_URL = PUSH_HOST + "/v3/push/validate"
	ALL       = "all"
	HTTP_POST = "POST"
)

var (
	ErrInvalidPlatform         = errors.New("<Platform>: invalid platform")
	ErrMessageContentMissing   = errors.New("<Message>: msg_content is required.")
	ErrContentMissing          = errors.New("<PushObject>: notification or message is required")
	ErrIosNotificationTooLarge = errors.New("<IosNotification>: iOS notification too large")
)

func NewJPushClient(appKey, masterSecret string) *JPushClient {
	client := JPushClient{
		appKey:       appKey,
		masterSecret: masterSecret,
	}
	headers := make(map[string]string)
	headers["User-Agent"] = USER_AGENT
	headers["Connection"] = CONNECTION_ALIVE
	headers["Charset"] = CHARSET
	headers["Content-Type"] = CONTENT_TYPE_JSON
	headers["Authorization"] = BasicAuth(appKey, masterSecret)
	client.headers = headers
	return &client
}

//推送
func (p *JPushClient) Push() ([]byte, error) {
	ret, err := PostJSON(p, PUSH_URL)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *JPushClient) SetPlayload(pi *Playload) {
	p.wrapper = pi
}

//playload 方法
func NewPlayload() *Playload {
	return &Playload{}
}

func NewPlatform() *Platform {
	return &Platform{}
}

func NewAudience() *Audience {
	return &Audience{}
}

func NewMessage(content string) *Message {
	return &Message{Content: content}
}

func NewNotification(alert string) *Notification {
	return &Notification{Alert: alert}
}

func NewOptions() *Options {
	return &Options{}
}

//设置推送平台[all,ios,android,winphone]
func (p *Platform) Set(platforms ...string) error {
	argSize := len(platforms)
	if argSize == 0 {
		return ErrInvalidPlatform
	} else if argSize == 1 && platforms[0] == "all" {
		p.isAll = true
		return nil
	}
	if p.value == nil {
		p.value = make([]string, 0)
	}
	for _, platform := range platforms {
		p.value = append(p.value, platform)
	}
	return nil
}

func (p *Platform) All() {
	p.isAll = true
}

func (p *Platform) Value() interface{} {
	if p.isAll {
		return ALL
	}
	return p.value
}

func (p *Audience) SetTag(tags []string) {
	p.set("tag", tags)
}

func (p *Audience) SetTagAnd(tagAnds []string) {
	p.set("tag_and", tagAnds)
}

func (p *Audience) SetAlias(alias []string) {
	p.set("alias", alias)
}

func (p *Audience) SetRegistrationId(ids []string) {
	p.set("registration_id", ids)
}

func (p *Audience) set(key string, v []string) {
	if p.value == nil {
		p.value = make(map[string][]string)
	}
	p.value[key] = v
}

func (p *Audience) All() {
	p.isAll = true
}

func (p *Audience) Value() interface{} {
	if p.isAll {
		return ALL
	}
	return p.value
}

//message -- 增加拓展信息
func (p *Message) AddExtra(value map[string]interface{}) {
	if p.Extras == nil {
		p.Extras = make(map[string]interface{})
	}
	p.Extras = value
}

//notification
func (p *platformNotification) AddExtra(value map[string]interface{}) {
	if p.Extras == nil {
		p.Extras = make(map[string]interface{})
	}
	p.Extras = value
}

func NewIosNotification(alert string) *IosNotification {
	p := &IosNotification{}
	p.Alert = alert
	return p
}

func NewAndroidNotification(alert string) *AndroidNotification {
	n := &AndroidNotification{}
	n.Alert = alert
	return n
}

//设备方法
func NewDeviceUpdate() *DeviceUpdate {
	return &DeviceUpdate{
		Tags: tags{},
	}
}

func (du *DeviceUpdate) MarshalJSON() ([]byte, error) {
	wrapper := deviceUpdateWrapper{}
	if du.Tags.Clear {
		wrapper.Tags = ""
	} else {
		wrapper.Tags = du.Tags
	}
	wrapper.Alias = du.Alias
	wrapper.Mobile = du.Mobile
	return json.Marshal(wrapper)
}

func (du *DeviceUpdate) AddTags(tags ...string) {
	du.Tags.Clear = false
	du.Tags.Add = append(du.Tags.Add, tags...)
}

func (du *DeviceUpdate) RemoveTags(tags ...string) {
	du.Tags.Clear = false
	du.Tags.Remove = append(du.Tags.Remove, tags...)
}

func (du *DeviceUpdate) ClearAllTags() {
	du.Tags.Clear = true
}

func (du *DeviceUpdate) SetAlias(alias string) {
	du.Alias = alias
}

func (du *DeviceUpdate) SetMobile(mobile string) {
	du.Mobile = mobile
}
