package mqttbinding

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNilSchema(t *testing.T) {
	logrus.Infof("--- TestNilSchema ---")
	data1 := "text"

	io := NewInteractionOutput(data1, nil)

	asValue := io.Value()
	assert.Equal(t, data1, asValue)

}

func TestArray(t *testing.T) {
	data1 := []string{"item 1", "item 2"}
	io := NewInteractionOutput(data1, nil)
	asArray := io.ValueAsArray()
	assert.Len(t, asArray, 2)
}

func TestBool(t *testing.T) {
	data1 := true
	io := NewInteractionOutput(data1, nil)
	asBool := io.ValueAsBoolean()
	assert.Equal(t, true, asBool)
}

func TestInt(t *testing.T) {
	data1 := 42
	io := NewInteractionOutput(data1, nil)
	asInt := io.ValueAsInt()
	assert.Equal(t, 42, asInt)
}

func TestObject(t *testing.T) {
	type User struct {
		Name        string
		Age         int
		Active      bool
		LastLoginAt string
	}
	u1 := User{Name: "Bob", Age: 10, Active: true, LastLoginAt: "today"}
	io := NewInteractionOutput(u1, nil)
	asObject := io.ValueAsMap()
	assert.Equal(t, u1.Name, asObject["Name"])

}
