package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/am2b/tomlctl/internal/utils"
    "os"
    "strings"
)

func usage() {
    fmt.Fprintln(os.Stderr,
        `tomlctl: 脚本友好的 TOML 管理工具

用法:
    tomlctl get  <toml> <path>          - 读取值
    tomlctl set  <toml> <path> <value>  - 设置值 (自动创建路径)
    tomlctl has  <toml> <path>          - 检查路径是否存在 (Exit 0/1)
    tomlctl del  <toml> <path>          - 删除路径
    tomlctl list <toml> [path] [--json] - 列出指定路径下的内容

路径语法:
    a.b.c        (对象层级)
    a.b[0].c     (数组索引)

示例:
    tomlctl get config.toml personal.name
    tomlctl set config.toml personal.age 18
    tomlctl has config.toml personal.name
    tomlctl del config.toml personal.skills
    tomlctl list config.toml
    tomlctl list config.toml --json

注意:
    在zsh命令行中:包含数组索引的路径需要加引号,避免zsh解析[],在bash脚本中:无需
    tomlctl get config.toml "personal.contact[0]"
    tomlctl set config.toml "personal.skills.programming [bash, go]"
    tomlctl del config.toml "personal.skills.programming[0]"

另,遍历TOML数组表的万能模板:
    JSON=$(tomlctl list 配置文件 数组表路径 --json)
    echo "$JSON" | jq -c '.[]' | while read ELEM; do
        #提取字段
        字段1=$(echo "$ELEM" | jq -r '.字段1')
        字段2=$(echo "$ELEM" | jq -r '.字段2')
    done:`,
    )

    os.Exit(2)
}

func die(err error) {
    if err != nil {
        fmt.Fprintln(os.Stderr, "错误:", err)
        os.Exit(1)
    }
}

func requireArgs(n int) {
    if len(os.Args) != n {
        usage()
    }
}

// 命令封装
func cmdGet(file, path string) error {
    tree, err := utils.LoadFile(file)
    if err != nil {
        return err
    }

    v, ok := utils.Lookup(tree, path)
    if !ok {
        return errors.New("路径不存在")
    }

    fmt.Println(v)

    return nil
}

func cmdSet(file, path, raw string) error {
    tree, err := utils.LoadFile(file)
    if err != nil {
        return err
    }

    val := utils.InferValue(raw)

    if err := utils.SetValue(tree, path, val); err != nil {
        return err
    }

    return utils.SaveFile(file, tree)
}

func cmdDel(file, path string) error {
    tree, err := utils.LoadFile(file)
    if err != nil {
        return err
    }

    if err := utils.DeletePath(tree, path); err != nil {
        return err
    }

    return utils.SaveFile(file, tree)
}

func cmdHas(file, path string) error {
    tree, err := utils.LoadFile(file)
    if err != nil {
        return err
    }

    _, ok := utils.Lookup(tree, path)
    if ok {
        os.Exit(0)
    }
    os.Exit(1)

    return nil
}

func cmdList(args []string) error {
    if len(args) < 1 {
        return errors.New("参数不足:至少需要指定TOML文件路径")
    }

    file := args[0]

    path, jsonOut := "", false
    // 统计路径参数数量(只能有0或1个)
    pathCount := 0

    for _, a := range args[1:] {
        // 排除空参数(防御用户误输入的空字符串)
        if strings.TrimSpace(a) == "" {
            return errors.New("无效参数:空字符串参数")
        }

        if a == "--json" {
            jsonOut = true
            continue
        }

        // 路径参数:只能有1个,多余则报错
        pathCount++
        if pathCount > 1 {
            return fmt.Errorf("参数错误:只能指定一个路径参数(当前输入了多个:%s,%s...)", path, a)
        }
        path = a
    }

    tree, err := utils.LoadFile(file)
    if err != nil {
        return err
    }

    var v any = tree
    if path != "" {
        vv, ok := utils.Lookup(tree, path)
        if !ok {
            return errors.New("路径不存在")
        }
        v = vv
    }

    if jsonOut {
        enc := json.NewEncoder(os.Stdout)
        enc.SetIndent("", "  ")
        if err := enc.Encode(v); err != nil {
            return fmt.Errorf("JSON 序列化失败:%w", err)
        }
    } else {
        fmt.Printf("%v\n", v)
    }

    return nil
}

func main() {
    if len(os.Args) < 2 {
        usage()
    }

    cmd := os.Args[1]
    switch cmd {
    case "get":
        requireArgs(4)
        die(cmdGet(os.Args[2], os.Args[3]))
    case "set":
        requireArgs(5)
        die(cmdSet(os.Args[2], os.Args[3], os.Args[4]))
    case "has":
        requireArgs(4)
        die(cmdHas(os.Args[2], os.Args[3]))
    case "del":
        requireArgs(4)
        die(cmdDel(os.Args[2], os.Args[3]))
    case "list":
        die(cmdList(os.Args[2:]))
    default:
        usage()
    }
}
