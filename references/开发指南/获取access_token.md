# 获取 access_token

> 最后更新：2024/03/26

# 注意事项

⚠️ 为了安全考虑：

开发者 **请勿** 将 `access_token` 返回给前端。

正确做法：

* `access_token` 应保存在服务端后台
* 所有企业微信 API 请求应由后台发起

---

# access_token 说明

获取 `access_token` 是调用企业微信 API 的第一步。

它相当于创建了一个登录凭证，后续所有业务 API：

* 都依赖 `access_token`
* 用于鉴权调用者身份

因此开发者在调用业务接口前，需要明确：

* `access_token` 的来源
* 使用正确应用对应的 `access_token`

---

# 接口信息

## 请求方式

```text id="c0v8fd"
GET（HTTPS）
```

---

## 请求地址

```text id="q8g4tr"
https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=ID&corpsecret=SECRET
```

---

# 提示

请求地址中的：

* `ID`
* `SECRET`

均为需要替换的变量。

其它接口也采用相同规则，不再重复说明。

---

# 参数说明

| 参数           | 必须 | 说明                                |
| ------------ | -- | --------------------------------- |
| `corpid`     | 是  | 企业 ID，参考：术语说明 → corpid            |
| `corpsecret` | 是  | 应用凭证密钥，应用必须为启用状态，参考：术语说明 → secret |

---

# 权限说明

每个应用都有独立的 `secret`。

因此：

* 每个应用获取的 `access_token`
* 只能用于当前应用

所以：

✅ 不同应用必须分别获取并缓存自己的 `access_token`

---

# 返回结果

```json id="a3pm2r"
{
  "errcode": 0,
  "errmsg": "ok",
  "access_token": "accesstoken000001",
  "expires_in": 7200
}
```

---

# 返回参数说明

| 参数             | 说明                      |
| -------------- | ----------------------- |
| `errcode`      | 错误码，`0` 表示成功，非 `0` 表示失败 |
| `errmsg`       | 返回码提示信息                 |
| `access_token` | 获取到的凭证，最长 512 字节        |
| `expires_in`   | 凭证有效时间（秒）               |

---

# 注意事项

## 缓存 access_token

开发者需要缓存 `access_token`，用于后续接口调用。

⚠️ 不要频繁调用 `gettoken` 接口，否则会触发频率限制。

---

## token 过期处理

当 `access_token`：

* 失效
* 过期

时，需要重新获取。

---

## 有效期说明

`access_token` 的有效期由 `expires_in` 返回。

正常情况下：

```text id="z1bsl4"
7200 秒（2 小时）
```

---

## 多应用缓存隔离

由于企业微信每个应用的 `access_token` 都是独立的：

开发者在缓存时，需要按应用分别存储。

---

## 存储空间要求

`access_token` 至少需要预留：

```text id="5tq1d8"
512 字节
```

的存储空间。

---

## 提前失效说明

企业微信可能因运营需要提前使 `access_token` 失效。

因此开发者必须实现：

```text id="p7e1kl"
access_token 失效后自动重新获取
```

的逻辑。
