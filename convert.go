package freetds

import (
  "errors"
  "strings"
  "bytes"
  "encoding/binary"
  "time"
	"fmt"
)

/*
#include <sybfront.h>
#include <sybdb.h>
*/
import "C"

const (
	//name               database type   go type
	SYBINT1 = 48       //tinyint       uint8
	SYBINT2 = 52       //smallint      int16
	SYBINT4 = 56       //int           int32
	SYBINT8 = 127      //bigint        int64

	SYBCHAR = 47
	SYBVARCHAR = 39    //nvarchar      string
	SYBNVARCHAR = 103  //nvarchar      string

	SYBREAL = 59       //real          float32
	SYBFLT8 = 62       //float(53)     float64
	SYBBIT = 50        //bit           bool

	SYBMONEY4 = 122    //smallmoney    float64
	SYBMONEY = 60      //money         float64

	SYBDATETIME = 61   //datetime      time.Time
	SYBDATETIME4 = 58  //smalldatetime time.Time

	SYBIMAGE = 34      //image         []byte
	SYBBINARY = 45     //binary        []byte
	SYBVARBINARY = 37  //varbinary     []byte
	XSYBVARBINARY = 165//varbinary     []byte
)

var sqlStartTime = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

func sqlBufToType(datatype int, data []byte) interface{} {
	buf := bytes.NewBuffer(data)
  switch datatype {
  case C.SYBINT1:
    var value uint8
    binary.Read(buf, binary.LittleEndian, &value)
    return value
  case C.SYBINT2:
    var value int16
    binary.Read(buf, binary.LittleEndian, &value)
    return value
  case C.SYBINT4:
    var value int32
    binary.Read(buf, binary.LittleEndian, &value)
    return value
  case C.SYBINT8:
    var value int64
    binary.Read(buf, binary.LittleEndian, &value)
    return value
  case C.SYBDATETIME:
    var days int32  /* number of days since 1/1/1900 */
    var sec  uint32 /* 300ths of a second since midnight */
    binary.Read(buf, binary.LittleEndian, &days)
    binary.Read(buf, binary.LittleEndian, &sec)
    value := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
    value = value.Add(time.Duration(days) * time.Hour * 24).Add(time.Duration(sec) * time.Second / 300)
    return value
  case C.SYBDATETIME4:
    var days uint16  /* number of days since 1/1/1900 */
    var mins  uint16 /* number of minutes since midnight */
    binary.Read(buf, binary.LittleEndian, &days)
    binary.Read(buf, binary.LittleEndian, &mins)
    value := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
    value = value.Add(time.Duration(days) * time.Hour * 24).Add(time.Duration(mins) * time.Minute)
    return value
  case C.SYBMONEY:
    var high int32
    var low  uint32
    binary.Read(buf, binary.LittleEndian, &high)
    binary.Read(buf, binary.LittleEndian, &low)
    return float64(int64(high) * 4294967296 + int64(low)) / 10000
  case C.SYBMONEY4 :
    var value int32
    binary.Read(buf, binary.LittleEndian, &value)
    return float64(value) / 10000
  case C.SYBREAL:
    var value float32
    binary.Read(buf, binary.LittleEndian, &value)
    return value
  case C.SYBFLT8:
    var value float64
    binary.Read(buf, binary.LittleEndian, &value)
    return value
  case C.SYBBIT:
    return data[0] == 1
  case C.SYBIMAGE, C.SYBVARBINARY, C.SYBBINARY: 
    return append([]byte{},  data[:len(data)-1]...) // make copy of data
    //TODO - decimal & numeric datatypes
	default: //string
		len := strings.Index(string(data), "\x00")
    return string(data[:len])
  }
}

func typeToSqlBuf(datatype int, value interface{}) (data []byte, err error) {
	buf := new(bytes.Buffer)
	switch datatype {
	case C.SYBINT1: { 
		if typedValue, ok := value.(uint8); ok {
			err = binary.Write(buf, binary.LittleEndian, typedValue)
		} else {
			err = errors.New(fmt.Sprintf("Could not convert %T to uint8.", value))
		}
	}
	case C.SYBINT2: { 
		if typedValue, ok := value.(int16); ok {
			err = binary.Write(buf, binary.LittleEndian, typedValue)
		} else {
			err = errors.New(fmt.Sprintf("Could not convert %T to int16.", value))
		}
	}
	case C.SYBINT4: {
		var int32Value int32
		switch value.(type) { 
		case int: {
			intValue, _ := value.(int)
			int32Value = int32(intValue)
		}
		case int32: 
			int32Value, _ = value.(int32)
		case int64: 
			intValue, _ := value.(int64)
			int32Value = int32(intValue)
		default: {
			err = errors.New(fmt.Sprintf("Could not convert %T to int32.", value))
			return
		}
		}
		err = binary.Write(buf, binary.LittleEndian, int32Value)
	}
	case C.SYBINT8: { 
		if typedValue, ok := value.(int64); ok {
			err = binary.Write(buf, binary.LittleEndian, typedValue)
		} else {
			err = errors.New(fmt.Sprintf("Could not convert %T to int64.", value))
		}
	}
	case C.SYBREAL: { 
		if typedValue, ok := value.(float32); ok {
			err = binary.Write(buf, binary.LittleEndian, typedValue)
		} else {
			err = errors.New(fmt.Sprintf("Could not convert %T to float32.", value))
		}
	}
	case C.SYBFLT8: { 
		if typedValue, ok := value.(float64); ok {
			err = binary.Write(buf, binary.LittleEndian, typedValue)
		} else {
			err = errors.New(fmt.Sprintf("Could not convert %T to float64.", value))
		}
	}
	case C.SYBBIT:
		if typedValue, ok := value.(bool); ok {
			if typedValue {
				data = []byte{1}
			} else {
				data = []byte{0}
			}
			return
		} else {
			err = errors.New(fmt.Sprintf("Could not convert %T to bool.", value))
		}
	case C.SYBMONEY4:{
		if typedValue, ok := value.(float64); ok {
			intValue := int32(typedValue * 10000)
			err = binary.Write(buf, binary.LittleEndian, intValue)
		} else {
			err = errors.New(fmt.Sprintf("Could not convert %T to float64.", value))
		}
	}
	case C.SYBMONEY: {
		if typedValue, ok := value.(float64); ok {
			intValue := int64(typedValue * 10000)
			high := int32(intValue >> 32)
			low := uint32(intValue - int64(high))
			err = binary.Write(buf, binary.LittleEndian, high)
			if err == nil {
				err = binary.Write(buf, binary.LittleEndian, low)
			}
		} else {
			err = errors.New(fmt.Sprintf("Could not convert %T to float64.", value))
		}
	}
	case C.SYBDATETIME: {
		if typedValue, ok := value.(time.Time); ok {
			typedValue = typedValue.UTC()
			days := int32(typedValue.Sub(sqlStartTime).Hours() / 24)
			secs := uint32((
				((typedValue.Hour() * 60 + typedValue.Minute()) * 60) + typedValue.Second()) * 300 +
					typedValue.Nanosecond() / 3333333)
			err = binary.Write(buf, binary.LittleEndian, days)
			if err == nil {
				err = binary.Write(buf, binary.LittleEndian, secs)
			}
		} else {
			err = errors.New(fmt.Sprintf("Could not convert %T to time.Time.", value))
		}
	}
	case C.SYBDATETIME4: {
		if typedValue, ok := value.(time.Time); ok {
			typedValue = typedValue.UTC()
			days := uint16(typedValue.Sub(sqlStartTime).Hours() / 24)
			mins := uint16(typedValue.Hour() * 60 + typedValue.Minute())
			err = binary.Write(buf, binary.LittleEndian, days)
			if err == nil {
				err = binary.Write(buf, binary.LittleEndian, mins)
			}
		} else {
			err = errors.New(fmt.Sprintf("Could not convert %T to time.Time.", value))
		}
	}
	case C.SYBIMAGE, C.SYBVARBINARY, C.SYBBINARY, XSYBVARBINARY:
		if typedValue, ok := value.([]byte); ok {
			data = append(typedValue, []byte{0}[0])
			return
		} else {
			err = errors.New(fmt.Sprintf("Could not convert %T to []byte.", value))
		}
	default: {
		if typedValue, ok := value.(string); ok {
			data = append([]byte(typedValue), []byte{0}[0])
			return
		} else {
			err = errors.New(fmt.Sprintf("Could not convert %T to string.", value))
		}
	}
	}
	data = buf.Bytes()
	return
}

func dbbindtype(datatype C.int) C.int {
  switch datatype {
  case C.SYBIMAGE, C.SYBVARBINARY, C.SYBBINARY:
    return C.BINARYBIND;
  case C.SYBBIT:
    return C.BITBIND;
  case C.SYBTEXT, C.SYBVARCHAR, C.SYBCHAR:
    return C.NTBSTRINGBIND;
  case C.SYBDATETIME:
    return C.DATETIMEBIND;
  case C.SYBDATETIME4:
    return C.SMALLDATETIMEBIND;
  case C.SYBDECIMAL:
    return C.DECIMALBIND;
  case C.SYBNUMERIC:
    return C.NUMERICBIND;
  case C.SYBFLT8:
    return C.FLT8BIND;
  case C.SYBREAL:
    return C.REALBIND;
  case C.SYBINT1:
    return C.TINYBIND;
  case C.SYBINT2:
    return C.SMALLBIND;
  case C.SYBINT4:
    return C.INTBIND;
  case C.SYBINT8:
    return C.BIGINTBIND;
  case C.SYBMONEY:
    return C.MONEYBIND;
  case C.SYBMONEY4:
    return C.SMALLMONEYBIND;
  }
  //TODO - log unknown datatype
  return C.NTBSTRINGBIND;
}
