# 获取成员会议 ID 列表

**最后更新：2026/03/23**

该接口用于获取指定成员在指定时间范围内的会议 ID 列表。

---

## 接口信息

* **请求方式**：`POST`（HTTPS）
* **请求地址**：

```text id="2r7vka"
https://qyapi.weixin.qq.com/cgi-bin/meeting/get_user_meetingid?access_token=ACCESS_TOKEN
```

---

## 请求包体

```json id="w6md8n"
{
  "userid": "USERID",
  "cursor": "cursor",
  "begin_time": 1586136317,
  "end_time": 1586236317,
  "limit": 100
}
```

---

# 参数说明

| 参数             | 必须 | 类型       | 说明                                                                                  |
| -------------- | -- | -------- | ----------------------------------------------------------------------------------- |
| `access_token` | 是  | `string` | 调用接口凭证。获取方法查看“获取 access_token”                                                      |
| `userid`       | 是  | `string` | 企业成员的 userid                                                                        |
| `cursor`       | 否  | `string` | 上一次调用返回的 cursor，首次调用可填 `"0"`                                                        |
| `limit`        | 否  | `uint32` | 每次拉取的数据量，默认值和最大值均为 `100`                                                            |
| `begin_time`   | 否  | `uint32` | 开始时间（Unix 时间戳）                                                                      |
| `end_time`     | 否  | `uint32` | 结束时间（Unix 时间戳），时间跨度不超过 180 天。如果 `begin_time` 和 `end_time` 都未填写，则默认 `end_time` 为当前时间 |

---

# 权限说明

* 只能拉取当前应用创建的会议 ID
* 自建应用需要配置在“可调用接口的应用”列表中

---

# 返回结果

```json id="k3u4cs"
{
  "errcode": 0,
  "errmsg": "ok",
  "next_cursor": "1223",
  "meetingid_list": [
    "meetingid1",
    "meetingid2"
  ]
}
```

---

# 返回参数说明

| 参数               | 类型         | 说明                                                   |
| ---------------- | ---------- | ---------------------------------------------------- |
| `errcode`        | `int32`    | 返回码                                                  |
| `errmsg`         | `string`   | 返回码描述内容                                              |
| `next_cursor`    | `string`   | 当前数据最后一个 key 值。下次调用时传入该值可继续分页拉取。若未返回或为空字符串，表示数据已拉取完毕 |
| `meetingid_list` | `string[]` | 会议 ID 列表，可能为空                                        |
