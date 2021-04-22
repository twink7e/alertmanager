# Alertmanger
>这个项目fork from https://github.com/prometheus/alertmanager
>在此基础上支持了云片的短信（不能超过1000个字符）和电话告警



## Build 

```
$ make build
```

## Config
alertmanager config file exmpale
```
global:
  resolve_timeout: 5m
  http_config:
    follow_redirects: true
route:
  receiver: sendcall
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 30m
  group_by:
  - alertname
  routes:
  - receiver: sendcall
    group_wait: 10s
    group_by:
    - alertname
    continue: true
inhibit_rules:
- source_match:
    severity: critical
  target_match:
    severity: warning
  equal:
  - alertname
  - dev
  - instance
receivers:
- name: sendcall
  yunpian_sendcall_configs:
  - send_resolved: false
    apikey: your yunpian api key
    mobile: phone num
    code: <code>
# 使用默认的告警模板，告警内容超出1k字会发送短信失败（云片接口会报错）
- name: sendsms
  yunpian_sendsms_configs:
  - send_resolved: false
    apikey: your yunpian api key
    mobile: phone num
templates: []
```