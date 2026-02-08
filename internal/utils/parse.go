package utils

import (
    "fmt"
    "strconv"
    "strings"
)

// 路径解析
type pathToken struct {
    Key   string
    Index int
    IsIdx bool
}

// 将path:"a.b[0].c"转换为Token序列
func ParsePath(path string) ([]pathToken, error) {
    var tokens []pathToken
    // 首先按点切分
    parts := strings.Split(path, ".")
    for _, part := range parts {
        // 处理用户不小心输入a..b这种情况
        if part == "" {
            continue
        }

        // 处理带索引的key,如:tags[0]
        if strings.Contains(part, "[") {
            // 找到[和]的位置
            openIdx := strings.Index(part, "[")
            closeIdx := strings.Index(part, "]")

            // 防止用户输入a[0或a]0[
            if closeIdx == -1 || closeIdx < openIdx {
                return nil, fmt.Errorf("路径语法错误: 缺失结束括号 ']' 在 %s", part)
            }

            // 提取key部分,b[0] -> b
            key := part[:openIdx]
            if key != "" {
                // 把b作为普通key加入token
                tokens = append(tokens, pathToken{Key: key})
            }

            // 提取index部分,[0] -> 0
            idxStr := part[openIdx+1 : closeIdx]
            idx, err := strconv.Atoi(idxStr)
            if err != nil {
                return nil, fmt.Errorf("非法数组索引: %s", idxStr)
            }
            // 加入一个IsIdx:true的token
            tokens = append(tokens, pathToken{Index: idx, IsIdx: true})

            // 处理索引后还有多余字符的情况(如:tags[0]xx)
            if closeIdx+1 < len(part) {
                // 如果后面还有内容且不是以点开始,递归处理或简单报错
                return nil, fmt.Errorf("不支持的路径格式: %s", part)
            }
        } else { // 普通段(不带[) -> 直接当key
            tokens = append(tokens, pathToken{Key: part})
        }
    }
    return tokens, nil
}
