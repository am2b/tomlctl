package utils

import (
    "errors"
)

// 输入:
// TOML解析后的根对象tree(map[string]any)
// 路径字符串(如a.b[0].c)
// 输出:找到的值+是否存在(any,bool)
func Lookup(tree map[string]any, path string) (any, bool) {
    // 先把路径解析成token
    tokens, err := ParsePath(path)
    if err != nil || len(tokens) == 0 {
        return nil, false
    }

    // cur:当前节点,初始指向整个TOML根对象
    // 类型:any(空接口),因为它可能是:
    // map[string]any(对象/表)
    // []any(数组)
    // 字符串/数字/布尔(叶子值)
    var cur any = tree
    for _, t := range tokens {
        // 情况A:普通键(非数组,IsIdx: false)
        if !t.IsIdx {
            m, ok := cur.(map[string]any)
            if !ok {
                return nil, false
            }
            cur, ok = m[t.Key]
            if !ok {
                return nil, false
            }
        } else { // 情况B:数组索引(IsIdx:true)
            a, ok := cur.([]any)
            if !ok || t.Index < 0 || t.Index >= len(a) {
                return nil, false
            }
            cur = a[t.Index]
        }
    }

    return cur, true
}

// SetValue:向TOML数据结构中设置指定路径的值,自动创建中间层级,支持数组自动扩容
// 输入:
// tree:TOML解析后的根map(map[string]any)
// path:要设置的路径(如:a.b[0].c)
// val:要设置的值(已经通过InferValue推断过类型)
// 输出:
// error(路径语法错误/类型不匹配)
func SetValue(tree map[string]any, path string, val any) error {
    tokens, err := ParsePath(path)
    if err != nil {
        return err
    }
    if len(tokens) == 0 {
        return errors.New("路径不能为空")
    }

    // 定义父节点结构体
    type parentNode struct {
        container any    // 父容器(map或数组)
        key       string // 父容器是map时的键
        idx       int    // 父容器是数组时的索引
        isArray   bool   // 父容器是否是数组
    }
    // 父节点链
    var parents []parentNode

    var cur any = tree
    for i := 0; i < len(tokens); i++ {
        t := tokens[i]
        isLast := i == len(tokens)-1

        if !t.IsIdx {
            // 处理普通键
            m, ok := cur.(map[string]any)
            if !ok {
                return errors.New("非 map 类型无法通过普通键访问")
            }

            // 记录父节点(非最后一个token时记录)
            if !isLast {
                parents = append(parents, parentNode{
                    container: cur,
                    key:       t.Key,
                    isArray:   false,
                })
            }

            if isLast {
                // 最后一个token:直接设值
                m[t.Key] = val
                return nil
            }

            // 自动创建中间层级
            next, ok := m[t.Key]
            if !ok || next == nil {
                if tokens[i+1].IsIdx {
                    m[t.Key] = []any{}
                } else {
                    m[t.Key] = map[string]any{}
                }
                next = m[t.Key]
            }
            cur = next
        } else {
            // 处理数组索引(支持自动扩容)
            a, ok := cur.([]any)
            if !ok {
                return errors.New("尝试对非数组定位索引")
            }

            // 记录父节点(非最后一个token时记录)
            if !isLast {
                parents = append(parents, parentNode{
                    container: cur,
                    idx:       t.Index,
                    isArray:   true,
                })
            }

            // 自动扩容逻辑
            needExpand := t.Index >= len(a)
            var newA []any
            if needExpand {
                // 扩容数组到目标索引+1的长度
                newA = make([]any, t.Index+1)
                // 复制原数组内容,扩容部分默认nil
                copy(newA, a)
            } else {
                // 无需扩容,用原数组
                newA = a
            }

            if isLast {
                // 最后一个token:设值(扩容后写回父容器)
                newA[t.Index] = val
                // 扩容后需要写回父容器
                if needExpand && len(parents) > 0 {
                    p := parents[len(parents)-1]
                    if p.isArray {
                        // 父容器是数组
                        parentA := p.container.([]any)
                        parentA[p.idx] = newA
                    } else {
                        // 父容器是map(绝大多数场景)
                        parentM := p.container.(map[string]any)
                        parentM[p.key] = newA
                    }
                }
                return nil
            }

            // 非最后一个token:处理中间层级
            if needExpand {
                // 扩容后初始化当前索引的子层级
                if tokens[i+1].IsIdx {
                    newA[t.Index] = []any{}
                } else {
                    newA[t.Index] = map[string]any{}
                }
                // 写回父容器
                if len(parents) > 0 {
                    p := parents[len(parents)-1]
                    if p.isArray {
                        parentA := p.container.([]any)
                        parentA[p.idx] = newA
                    } else {
                        parentM := p.container.(map[string]any)
                        parentM[p.key] = newA
                    }
                }
                cur = newA[t.Index]
            } else {
                // 无需扩容:初始化空的子层级
                if newA[t.Index] == nil {
                    if tokens[i+1].IsIdx {
                        newA[t.Index] = []any{}
                    } else {
                        newA[t.Index] = map[string]any{}
                    }
                }
                cur = newA[t.Index]
            }
        }
    }
    return nil
}

// 从TOML数据结构中删除指定路径的节点
// 输入:
// tree:TOML解析后的根map(map[string]any)
// path:要删除的路径(如:a.b[0].c)
// 输出:
// error(仅路径语法错误时返回,路径不存在/类型不匹配时不返回错误)
func DeletePath(tree map[string]any, path string) error {
    tokens, err := ParsePath(path)
    if err != nil {
        return err
    }
    if len(tokens) == 0 {
        return nil
    }

    // 记录父节点信息
    type parentNode struct {
        container any    // 父容器(map或数组)
        key       string // 父容器是map时的键
        idx       int    // 父容器是数组时的索引
        isArray   bool   // 父容器是否是数组
    }
    // 父节点链
    var parents []parentNode

    var cur any = tree
    for i := 0; i < len(tokens); i++ {
        t := tokens[i]
        isLast := i == len(tokens)-1

        if isLast {
            // 最后一个token:执行删除操作
            if !t.IsIdx {
                // 删除map键(原逻辑不变)
                m, ok := cur.(map[string]any)
                if ok {
                    delete(m, t.Key)
                }
            } else {
                // 删除数组元素
                a, ok := cur.([]any)
                if !ok || t.Index < 0 || t.Index >= len(a) {
                    return nil
                }
                // 1.真正删除元素:缩容切片
                newA := append(a[:t.Index], a[t.Index+1:]...)
                // 2.找到父节点,把缩容后的数组写回父容器
                if len(parents) > 0 {
                    p := parents[len(parents)-1]
                    if p.isArray {
                        // 父容器是数组
                        parentA := p.container.([]any)
                        parentA[p.idx] = newA
                    } else {
                        // 父容器是map
                        parentM := p.container.(map[string]any)
                        parentM[p.key] = newA
                    }
                }
            }
            return nil
        }

        // 非最后一个token:记录父节点,导航到下一级
        parents = append(parents, parentNode{
            container: cur,
            key:       t.Key,
            idx:       t.Index,
            isArray:   t.IsIdx,
        })

        if !t.IsIdx {
            m, ok := cur.(map[string]any)
            if !ok {
                return nil
            }
            cur = m[t.Key]
        } else {
            a, ok := cur.([]any)
            if !ok || t.Index >= len(a) {
                return nil
            }
            cur = a[t.Index]
        }
    }
    return nil
}
