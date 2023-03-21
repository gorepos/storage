## storage
Simple but awesome key-value file-based storage for Go (Golang). It allows to save/restore any data (strings, maps, structs etc) by string key

### Install

```bash
go get github.com/gorepos/storage
```

### Example

```Go
import "github.com/gorepos/storage"
func main() {
  // save string
  storage.Put("some-key1", "some value")
  
  // save map 
  storage.Put("some-key2", map[string]any{"Integer": 123, "String": "hello string"})
  
  
}
```
