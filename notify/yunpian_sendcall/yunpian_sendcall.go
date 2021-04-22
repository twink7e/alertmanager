// Copyright 2019 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yunpian_sendcall

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
	commoncfg "github.com/prometheus/common/config"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

var send_call_url = "https://voice.yunpian.com/v2/voice/send.json"

type yunpianSendCallResp struct {
	Count int    `json:"count"`
	Fee   int    `json:"fee"`
	Sid   string `json:"sid"`
}

type Notifier struct {
	conf   *config.YunpianSendCallConfig
	tmpl   *template.Template
	logger log.Logger
	client *http.Client
}

// New returns a new YunpianSendCall notifier.
func New(c *config.YunpianSendCallConfig, t *template.Template, l log.Logger) (*Notifier, error) {
	client, err := commoncfg.NewClientFromConfig(*c.HTTPConfig, "yunpian_sendcall", false, false)
	if err != nil {
		return nil, err
	}

	return &Notifier{conf: c, tmpl: t, logger: l, client: client}, nil
}

// Notify implements the Notifier interface.
func (n *Notifier) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	key, err := notify.ExtractGroupKey(ctx)
	if err != nil {
		return false, err
	}

	level.Debug(n.logger).Log("incident", key)
	data := notify.GetTemplateData(ctx, n.tmpl, as, n.logger)

	tmpl := notify.TmplText(n.tmpl, data, &err)
	if err != nil {
		return false, err
	}

	parameters := url.Values{}
	parameters.Add("apikey", tmpl(string(n.conf.APIkey)))
	parameters.Add("mobile", tmpl(n.conf.MobileNums))
	parameters.Add("code", tmpl(strconv.Itoa(n.conf.Code)))

	resp, err := n.client.PostForm(send_call_url, parameters)
	if err != nil {
		return true, notify.RedactURL(err)
	}
	defer notify.Drain(resp)

	if resp.StatusCode != 200 {
		return true, fmt.Errorf("unexpected status code %v", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return true, err
	}
	level.Debug(n.logger).Log("response", string(body), "incident", key)
	var ypSendCallResp yunpianSendCallResp
	if err := json.Unmarshal(body, &ypSendCallResp); err != nil {
		return true, err
	}
	if ypSendCallResp.Count > 1 {
		return false, nil
	}
	return false, errors.New(fmt.Sprint(ypSendCallResp))
}
