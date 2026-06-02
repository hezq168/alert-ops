package adapter

import "encoding/json"

// 使用标准库 JSON，方便后续替换为更高效的库
var jsonMarshal = json.Marshal
var jsonUnmarshal = json.Unmarshal
