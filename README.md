## storage
Simple but awesome key-value file-based storage for Go (Golang). It allows to save/restore any data (strings, maps, structs etc) by string key.

### Install

```bash
go get github.com/gorepos/storage
```

### Usage

```Go

import "github.com/gorepos/storage"


func main() {
  // define some structure
  type MyStruct struct {
    Name string
    Age  int
  }
  
  
  // save 
  storage.Put("some-key1", MyStruct {
    Name: "Ilon Mask"
    Age: 25
  })
  
  // restore 
  var restoredValue MyStruct
  storage.Get("some-key1", &restoredValue)
}
```
