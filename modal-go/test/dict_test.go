package test

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	pickle "github.com/kisielk/og-rek"
	"github.com/modal-labs/libmodal/modal-go"
	"github.com/onsi/gomega"
)

func TestDictInvalidName(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	for _, name := range []string{"has space", "has/slash", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"} {
		_, err := modal.DictLookup(context.Background(), name, nil)
		g.Expect(err).Should(gomega.HaveOccurred())
	}
}

func TestDictEphemeralBasicOperations(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	dict, err := modal.DictEphemeral(ctx, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()

	created, err := dict.Put("key1", "value1", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(created).To(gomega.BeTrue())

	value1, err := dict.Get("key1", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(value1).To(gomega.Equal("value1"))

	contains, err := dict.Contains("key1")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(contains).To(gomega.BeTrue())

	contains, err = dict.Contains("nonexistent")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(contains).To(gomega.BeFalse())

	var defaultVal any = "default"
	result, err := dict.Get("nonexistent", &defaultVal)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(result).To(gomega.Equal("default"))

	// Get non-existent key without default should return KeyError
	_, err = dict.Get("nonexistent", nil)
	g.Expect(errors.As(err, &modal.KeyError{})).To(gomega.BeTrue())

	// Get non-existent key with explicit nil default should return nil
	var nilDefault any = nil
	nilResult, err := dict.Get("nonexistent", &nilDefault)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(nilResult).To(gomega.BeNil())

	err = dict.Update(map[any]any{"key2": "value2", "key3": "value3"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	value2, err := dict.Get("key2", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(value2).To(gomega.Equal("value2"))

	value3, err := dict.Get("key3", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(value3).To(gomega.Equal("value3"))

	length, err := dict.Len()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(length).To(gomega.Equal(3))

	err = dict.Update(map[any]any{"key1": "newValue1", "key4": "value4"})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	newValue1, err := dict.Get("key1", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(newValue1).To(gomega.Equal("newValue1"))

	unchangedValue2, err := dict.Get("key2", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(unchangedValue2).To(gomega.Equal("value2"))

	value4, err := dict.Get("key4", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(value4).To(gomega.Equal("value4"))

	popped, err := dict.Pop("key1")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(popped).To(gomega.Equal("newValue1"))

	contains, err = dict.Contains("key1")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(contains).To(gomega.BeFalse())

	_, err = dict.Pop("nonexistent")
	g.Expect(errors.As(err, &modal.KeyError{})).To(gomega.BeTrue())

	err = dict.Clear()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	length, err = dict.Len()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(length).To(gomega.Equal(0))
}

func TestDictDifferentDataTypes(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	dict, err := modal.DictEphemeral(ctx, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()

	// Set up
	_, err = dict.Put("string", "hello", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = dict.Put("number", 42, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = dict.Put("boolean", true, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = dict.Put("object", map[string]any{"nested": "value"}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = dict.Put("array", []any{1, 2, 3}, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = dict.Put("null", nil, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	_, err = dict.Put(123, "numeric key", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	// Verify with Get
	stringVal, err := dict.Get("string", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(stringVal).To(gomega.Equal("hello"))

	numberVal, err := dict.Get("number", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(numberVal).To(gomega.Equal(int64(42)))

	boolVal, err := dict.Get("boolean", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(boolVal).To(gomega.Equal(true))

	objVal, err := dict.Get("object", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(objVal).To(gomega.Equal(map[any]any{"nested": "value"}))

	arrayVal, err := dict.Get("array", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(arrayVal).To(gomega.Equal([]any{int64(1), int64(2), int64(3)}))

	nullVal, err := dict.Get("null", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(nullVal).To(gomega.Equal(pickle.None{})) // Note: ogórek.None, not Go nil!

	numericKeyVal, err := dict.Get(123, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(numericKeyVal).To(gomega.Equal("numeric key"))

	// Verify with Pop
	stringVal, err = dict.Pop("string")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(stringVal).To(gomega.Equal("hello"))

	numberVal, err = dict.Pop("number")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(numberVal).To(gomega.Equal(int64(42)))

	boolVal, err = dict.Pop("boolean")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(boolVal).To(gomega.Equal(true))

	objVal, err = dict.Pop("object")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(objVal).To(gomega.Equal(map[any]any{"nested": "value"}))

	arrayVal, err = dict.Pop("array")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(arrayVal).To(gomega.Equal([]any{int64(1), int64(2), int64(3)}))

	nullVal, err = dict.Pop("null")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(nullVal).To(gomega.Equal(pickle.None{})) // Note: ogórek.None, not Go nil!

	numericKeyVal, err = dict.Pop(123)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(numericKeyVal).To(gomega.Equal("numeric key"))

	length, err := dict.Len()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(length).To(gomega.Equal(0))
}

func TestDictSkipIfExists(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	dict, err := modal.DictEphemeral(ctx, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()

	created, err := dict.Put("key", "value1", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(created).To(gomega.BeTrue())

	value, err := dict.Get("key", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(value).To(gomega.Equal("value1"))

	// Second Put without skipIfExists should overwrite
	created, err = dict.Put("key", "value2", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(created).To(gomega.BeTrue())

	value, err = dict.Get("key", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(value).To(gomega.Equal("value2"))

	// Put with skipIfExists should not overwrite
	created, err = dict.Put("key", "value3", &modal.DictPutOptions{SkipIfExists: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(created).To(gomega.BeFalse())

	value, err = dict.Get("key", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(value).To(gomega.Equal("value2"))

	// New key with skipIfExists should succeed
	created, err = dict.Put("newkey", "newvalue", &modal.DictPutOptions{SkipIfExists: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(created).To(gomega.BeTrue())

	value, err = dict.Get("newkey", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(value).To(gomega.Equal("newvalue"))
}

func TestDictIteration(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	dict, err := modal.DictEphemeral(ctx, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()

	testData := map[any]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	err = dict.Update(testData)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	keys := make([]any, 0)
	for key, err := range dict.Keys() {
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		keys = append(keys, key)
	}
	g.Expect(len(keys)).To(gomega.Equal(3))
	g.Expect(keys).To(gomega.ContainElements("key1", "key2", "key3"))

	values := make([]any, 0)
	for value, err := range dict.Values() {
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		values = append(values, value)
	}
	g.Expect(len(values)).To(gomega.Equal(3))
	g.Expect(values).To(gomega.ContainElements("value1", "value2", "value3"))

	items := make(map[any]any)
	for item, err := range dict.Items() {
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		items[item[0]] = item[1]
	}
	g.Expect(len(items)).To(gomega.Equal(3))
	g.Expect(items["key1"]).To(gomega.Equal("value1"))
	g.Expect(items["key2"]).To(gomega.Equal("value2"))
	g.Expect(items["key3"]).To(gomega.Equal("value3"))
}

func TestDictNonEphemeral(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	dictName := "test-dict-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	dict1, err := modal.DictLookup(ctx, dictName, &modal.LookupOptions{CreateIfMissing: true})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	defer func() {
		err := modal.DictDelete(ctx, dictName, nil)
		g.Expect(err).ShouldNot(gomega.HaveOccurred())

		_, err = modal.DictLookup(ctx, dictName, nil)
		g.Expect(err).Should(gomega.HaveOccurred())
	}()

	_, err = dict1.Put("persistent", "data", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	dict2, err := modal.DictLookup(ctx, dictName, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	value, err := dict2.Get("persistent", nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(value).To(gomega.Equal("data"))
}

func TestDictEmptyOperations(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	ctx := context.Background()

	dict, err := modal.DictEphemeral(ctx, nil)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	defer dict.CloseEphemeral()

	length, err := dict.Len()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(length).To(gomega.Equal(0))

	var defaultVal any = "default"
	value, err := dict.Get("key", &defaultVal)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(value).To(gomega.Equal("default"))

	contains, err := dict.Contains("key")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(contains).To(gomega.BeFalse())

	err = dict.Clear()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	err = dict.Update(map[any]any{})
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	length, err = dict.Len()
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(length).To(gomega.Equal(0))

	keys := make([]any, 0)
	for key, err := range dict.Keys() {
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		keys = append(keys, key)
	}
	g.Expect(keys).To(gomega.Equal([]any{}))
}
