# 获取企业微信接口 IP 段

> 最后更新：2023/04/24

## API 域名 IP 说明

API 域名 IP 指的是：

```text
qyapi.weixin.qq.com
```

的解析地址。

该地址是开发者调用企业微信端接口时使用的接入 IP。

---

## 使用场景

如果企业需要配置防火墙，可以通过该接口获取所有相关 IP 段。

注意事项：

* IP 段可能发生变更
* 当 IP 段变更时，新旧 IP 段会同时保留一段时间
* 建议企业每天定时拉取 IP 段
* 并更新防火墙设置
* 避免因 IP 段变更导致网络不通

---

# 接口信息

## 请求方式

```text
GET（HTTPS）
```

## 请求地址

```text
https://qyapi.weixin.qq.com/cgi-bin/get_api_domain_ip?access_token=ACCESS_TOKEN
```

---

# 请求参数说明

| 参数             | 必须 | 说明     |
| -------------- | -- | ------ |
| `access_token` | 是  | 调用接口凭证 |

---

# 权限说明

无限定。

---

# 返回结果

```json
{
  "ip_list": [
    "182.254.11.176",
    "182.254.78.66"
  ],
  "errcode": 0,
  "errmsg": "ok"
}
```

---

# 返回参数说明

| 参数        | 类型          | 说明                        |
| --------- | ----------- | ------------------------- |
| `ip_list` | StringArray | 企业微信服务器 IP 段              |
| `errcode` | int         | 错误码，`0` 表示成功，非 `0` 表示调用失败 |
| `errmsg`  | string      | 错误信息，调用失败时返回相关错误信息        |

---

# 调用失败判断

根据 `errcode` 判断调用是否失败：

```text
errcode != 0 表示调用失败
```

---

## access_token 过期返回示例

```json
{
  "ip_list": [],
  "errcode": 42001,
  "errmsg": "access_token expired, hint: [1576065934_28_e0fae07666aa64636023c1fa7e8f49a4], from ip: 9.30.0.138, more info at https://open.work.weixin.qq.com/devtool/query?e=42001"
}
```
