package storage

import (
	"fmt"
	. "github.com/onsi/gomega"
	"os"
	"strconv"
	"strings"
	"testing"
)

var testItems map[string]any = map[string]any{
	"aroot_key1":           map[string]any{"Integer": 123, "String": "hello string"},
	"subdir/innerkey":      map[string]any{"Integer": 123, "String": "hello inner key string"},
	"subdir/innerkey/just": "Just a string",
	"subdir2/hell/of/super/long/dir/inner/tree": []string{"one", "two", "three"},
	"orders/new/1000": []string{"one", "two", "three"},
	"orders/new/1001": []string{"one", "two", "three"},
	"orders/new/1003": []string{"one", "two", "three"},
	"orders/new/1004": map[string]any{
		"key1": 123,
		"key2": 0.333,
		"key3": "8888",
		"key4": []string{"one", "two", "three"},
	},
	"multiline": "line one\nline two\n    line three indented\nline four cyrillic Кирилица\n\n",
}

func TestInit(t *testing.T) {
	RegisterTestingT(t)
	testStorageDir := "test_storage_dir"
	os.RemoveAll(testStorageDir)
	SetOptions(Options{
		Dir: testStorageDir,
	})
}

func TestPut(t *testing.T) {
	RegisterTestingT(t)
	t.Logf("Initing storage..")
	//SetDirectory("./runtime_test")

	// put to storage
	for key, val := range testItems {
		err := Put(key, val)
		Expect(err).To(BeNil())
	}

	// check values
	for key, val := range testItems {
		var restoredValue any
		err := Get(key, &restoredValue)
		Expect(err).To(BeNil())
		t.Logf("restored %s", fmt.Sprint(restoredValue))
		t.Logf("orignval %s", fmt.Sprint(val))

		if fmt.Sprint(val) != fmt.Sprint(restoredValue) {
			t.Errorf("Not equal!!")
			t.Logf("restored %s", fmt.Sprint(restoredValue))
			t.Logf("orignval %s", fmt.Sprint(val))
			t.Fail()
		}
	}

}

func TestList(t *testing.T) {
	RegisterTestingT(t)
	allKeys := Keys("")
	// check all keys count
	Expect(len(allKeys)).To(Equal(len(testItems)))

	// check prefix filtering
	Expect(len(Keys("orders/new/"))).To(Equal(4))
}

func TestMove(t *testing.T) {
	keys1 := Keys("orders/new")
	for _, key := range keys1 {
		err := Move(key, "orders/done"+key[10:])
		Expect(err).To(BeNil())
	}

	Expect(len(Keys("orders/done"))).To(Equal(len(keys1)))

	// move back
	for _, key := range Keys("orders/done") {
		err := Move(key, "orders/new"+key[11:])
		Expect(err).To(BeNil())
	}
	Expect(len(Keys("orders/done"))).To(Equal(0))
	Expect(len(Keys("orders/new"))).To(Equal(len(keys1)))
}

func TestSlashes(t *testing.T) {
	RegisterTestingT(t)
	Put("////", "123")
	var loaded string
	Get("", &loaded)
	Expect(loaded).To(Equal("123"))
	Delete("")
}

func TestDelete(t *testing.T) {
	RegisterTestingT(t)
	Expect(len(Keys(""))).To(Equal(len(testItems)))
	for key, _ := range testItems {
		err := Delete(key)
		Expect(err).To(BeNil())
	}
	Expect(len(Keys(""))).To(Equal(0))
	Expect(isEmpty(gStorage.options.Dir)).To(BeTrue())
	err := os.Remove(gStorage.options.Dir)
	Expect(err).To(BeNil())
}

func TestMultiple(t *testing.T) {
	RegisterTestingT(t)

	storage1 := NewStorage(Options{
		Dir: "dir1",
	})
	storage2 := NewStorage(Options{
		Dir: "dir2",
	})

	err := storage1.Put("key", "Hello World!")
	Expect(err).To(BeNil())
	err = storage2.Put("key", "Hello Nether!")
	Expect(err).To(BeNil())

	err = storage1.Delete("key")
	Expect(err).To(BeNil())
	err = storage2.Delete("key")
	Expect(err).To(BeNil())

	os.Remove("dir1")
	os.Remove("dir2")
}

func TestForbiddenKeys(t *testing.T) {
	RegisterTestingT(t)

	badKeys := []string{
		"./",
		"./123",
		"../",
		"../123",

		"/.",
		"123/.",
		"/..",
		"123/..",

		"/../",
		"/./",
		"a/./b",
		"a/../b",
	}

	for _, key := range badKeys {
		err := Put(key, "val")
		if err != nil {
			t.Logf("checking key %-7s OK: \"%s\"", key, err.Error())
		}
		Expect(err).Should(HaveOccurred())
	}

}

func TestOptions(t *testing.T) {
	RegisterTestingT(t)

	// new instance
	s := NewStorage(Options{})
	Expect(s.options.Dir).Should(Equal(gDefaultOptions.Dir))
	s.SetOptions(Options{
		Dir: "somedir",
	})
	Expect(s.options.Dir).Should(Equal("somedir"))
	s.SetOptions(Options{
		Dir: "newdir",
	})
	Expect(s.options.Dir).Should(Equal("newdir"))

	// global
	oldDir := gStorage.options.Dir
	SetOptions(Options{Dir: "newdir"})
	Expect(gStorage.options.Dir).Should(Equal("newdir"))
	gStorage.options.Dir = oldDir
}

func BenchmarkStorage_Put(b *testing.B) {
	//b.Logf("N: %d", b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Put(strconv.Itoa(i), "123")
	}
	b.StopTimer()
	os.RemoveAll(gStorage.options.Dir)
}

func BenchmarkStorage_Put_Splitted(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := strings.Join(strings.Split(strconv.Itoa(i), ""), "/")
		Put(key, "123")
	}
	b.StopTimer()
	os.RemoveAll(gStorage.options.Dir)
}

func BenchmarkStorage_Get(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Put(strconv.Itoa(i), "123")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var val string
		Get(strconv.Itoa(i), &val)
	}
	b.StopTimer()
	os.RemoveAll(gStorage.options.Dir)
}

func BenchmarkStorage_Get_Splitted(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strings.Join(strings.Split(strconv.Itoa(i), ""), "/")
		Put(key, "123")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := strings.Join(strings.Split(strconv.Itoa(i), ""), "/")
		var val string
		Get(key, &val)
	}
	b.StopTimer()
	os.RemoveAll(gStorage.options.Dir)
}

func BenchmarkStorage_Move(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Put(strconv.Itoa(i), "123")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Move(strconv.Itoa(i), strconv.Itoa(i)+"_")
	}
	b.StopTimer()
	os.RemoveAll(gStorage.options.Dir)
}

func BenchmarkStorage_Delete(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Put(strconv.Itoa(i), "123")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Delete(strconv.Itoa(i))
	}
	b.StopTimer()
	os.RemoveAll(gStorage.options.Dir)
}
