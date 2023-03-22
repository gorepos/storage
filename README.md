## storage
Simple but awesome key-value file-based storage for Go (Golang). It allows to save/restore any data (strings, maps, structs etc) by string key.

**Features:**

- Each value is stored as seperated file (pretty-printed json)
- Key may contain slashes ("/") for example `users/123/orders/new/0001` and that value will be saved to the corresponding path `<storage-dir>/users/123/orders/new/0001.json`
- The ability to get a list of keys by prefix, for example `storage.Keys("users/123/orders/new")`

### Install

```bash
go get github.com/gorepos/storage
```

### Usage

```Go

import "github.com/gorepos/storage"

func main() {
  // save 
  storage.Put("some-key1", "Hello World!")
  
  // restore 
  var value string
  storage.Get("some-key1", &value)
}
```

### Save/Restore structs

```Go
  // define some structure
  type MyStruct struct {
    Name string
    Age  int
  }
  
  // save 
  storage.Put("some-key2", MyStruct {
    Name: "Elon Mask",
    Age:  25,
  })
  
  // restore 
  var restoredValue MyStruct
  storage.Get("some-key2", &restoredValue)

```

### Save/Restore maps

```Go
  // define some map
  var myMap = map[string]any{
    "name": "Elon Mask",
    "age": 25,
  }

  // save
  storage.Put("some-key3", myMap)

  // restore
  var restoredMap map[string]any
  storage.Get("some-key3", &restoredMap)

```
