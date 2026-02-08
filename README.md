# tomlctl
一个轻量、脚本友好的 TOML 命令行管理工具，支持读取、修改、删除、查询嵌套结构

---

## 功能一览
- get：读取指定路径的值
- set：设置值（自动创建中间层级）
- del：删除指定路径
- has：检查路径是否存在（通过退出码返回）
- list：查看节点内容，支持 JSON 格式化输出

---

## 使用示例

### 1. get — 读取值
tomlctl get config.toml server.port<br>
tomlctl get config.toml "server.tags[0]"

### 2. set — 设置值
tomlctl set config.toml server.port 8080<br>
tomlctl set config.toml server.tags "[dev, test, prod]"<br>
tomlctl set config.toml "server.tags[1]" staging

### 3. del — 删除值
tomlctl del config.toml server.port<br>
tomlctl del config.toml "server.tags[1]"<br>

### 4. has — 检查路径是否存在
tomlctl has config.toml server.port<br>
echo $?  # 0=存在，1=不存在

### 5. list — 查看结构（支持 JSON）
tomlctl list config.toml<br>
tomlctl list config.toml server --json

---

## 许可
MIT License
