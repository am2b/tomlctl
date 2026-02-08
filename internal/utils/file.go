package utils

import(
    "fmt"
    "os"
    "time"
    "path/filepath"
    "github.com/pelletier/go-toml/v2"
)

func LoadFile(path string) (map[string]any, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var tree map[string]any

    return tree, toml.Unmarshal(data, &tree)
}

func SaveFile(path string, tree map[string]any) error {
    out, err := toml.Marshal(tree)
    if err != nil {
        return err
    }

    dir := filepath.Dir(path)
    tmp := filepath.Join(dir, fmt.Sprintf(".%s.%d.tmp", filepath.Base(path), time.Now().UnixNano()))

    if err := os.WriteFile(tmp, out, 0644); err != nil {
        return err
    }

    return os.Rename(tmp, path)
}
