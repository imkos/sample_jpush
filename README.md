# **simple_jpush**



基于 极光官方文档【 https://docs.jiguang.cn/jpush/server/push/server_overview/ 】写的一个push lib.



Sample Code:

```go
package main

import (
	"jpush"
	"jsonc"
)

const (
	MAX_PUSH_WORKER = 5
)

var (
	ch_push_msgs = make(chan *st_jg_playload, 200)
	//
	push_opt = &jpush.Options{
		//最长10天
		TimeToLive: 10 * 60 * 60 * 24,
	}
)

type st_jg_push struct {
	*jpush.JPushClient
	plat *jpush.Platform
}

type st_jg_playload struct {
	recv_ids []string
	v_msg    interface{}
}

func NewPlayload(ids []string, v interface{}) *st_jg_playload {
	return &st_jg_playload{
		recv_ids: ids,
		v_msg:    v,
	}
}

func init_push_worker(appKey, masterSecret string) {
	android_plat := jpush.NewPlatform()
	android_plat.Set("android")
	//初始工作
	for i := 0; i < MAX_PUSH_WORKER; i++ {
		jpush := &st_jg_push{jpush.NewJPushClient(appKey, masterSecret), android_plat}
		go jpush.DoWork()
	}
}

func (p *st_jg_push) DoWork() {
	for {
		select {
		case p_msg := <-ch_push_msgs:
			//audience：推送目标
			au := jpush.NewAudience()
			au.SetAlias(p_msg.recv_ids)
			//notification：通知
			//message：自定义消息
			pmsg := &jpush.Message{
				Title:       "******",
				Content:     "Success",
				ContentType: "PAY",
				Extras:      p_msg.v_msg,
			}
			jplayload := &jpush.Playload{
				Platform: p.plat.Value(),
				Audience: au.Value(),
				Message:  pmsg,
				Options:  push_opt,
			}
			p.SetPlayload(jplayload)
			b_resp, err := p.Push()
			if err != nil {
				b_json, _ := json.Marshal(jplayload)
				Slog_Error("jg Push fail!", err, string(b_json))
				//跳出当前的select
				break
			}
			//
			Slog_Op(string(b_resp), err)
		}
	}
}
```

