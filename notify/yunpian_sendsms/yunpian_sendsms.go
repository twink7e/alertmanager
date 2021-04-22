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

package yunpian_sendsms

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
)

var send_sms_url = "https://sms.yunpian.com/v2/sms/single_send.json"

type yunpianSendSMSResp struct {
	Code   int     `json:"code"`
	Msg    string  `json:"msg"`
	Count  int     `json:"count,omitempty"`
	Fee    float64 `json:"fee,omitempty"`
	Unit   string  `json:"unit,omitempty"`
	Mobile string  `json:"moble,omitempty"`
	Sid    int64   `json:"sid,omitempty"`
}

type Notifier struct {
	conf   *config.YunpianSendSMSConfig
	tmpl   *template.Template
	logger log.Logger
	client *http.Client
}

// New returns a new YunpianSendSMS notifier.
func New(c *config.YunpianSendSMSConfig, t *template.Template, l log.Logger) (*Notifier, error) {
	client, err := commoncfg.NewClientFromConfig(*c.HTTPConfig, "yunpian_sendsms", false, false)
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
	parameters.Add("mobile", tmpl(n.conf.Mobile))
	parameters.Add("text", tmpl(n.conf.Text))
	resp, err := n.client.PostForm(send_sms_url, parameters)
	if err != nil {
		return true, notify.RedactURL(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer notify.Drain(resp)

	if resp.StatusCode != 200 {
		return true, fmt.Errorf("unexpected status code %v, response body: %s", resp.StatusCode, body)
	}
	if err != nil {
		return true, err
	}
	level.Debug(n.logger).Log("response", string(body), "incident", key)
	var ypSendSMSResp yunpianSendSMSResp
	if err := json.Unmarshal(body, &ypSendSMSResp); err != nil {
		return true, err
	}
	if ypSendSMSResp.Code > 0 {
		return false, nil
	}
	return false, errors.New(ypSendSMSResp.Msg)
}
