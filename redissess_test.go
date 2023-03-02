package redissess

import (
	"testing"
	// "fmt"
	"net/http"
)

var rclient *RedisClient


func createClient() *RedisClient {

	coninfo := ConnInfo{
		Addr:"172.18.0.201:6379",
		DB: 0,
	}

	redislifetime := 3600

	sess := SessionParam{
		Name:"ss",
		Path:"/",
		HttpOnly:true,
		Secure:true,	
		MaxAge:0,
		SameSite:http.SameSiteNoneMode,
		Lifetime:&redislifetime,
	}

	return New(coninfo,sess)

	
}

type TestObject struct {
	Name string
	Value string
}



func TestX(t *testing.T)  {
	
	rclient := createClient()

	if rclient==nil {
		t.Error("not connected")
		return
	}

	defer func(){
		t.Log("close")
		rclient.Close()
	}()

	suffix1 := "suffix1"
	suffix2 := "suffix2"


	// cookie,err := rclient.CreateAndSet(suffix1,TestObject{"testname","testvalue"})

	// if err != nil {
	// 	t.Error(err)

	// 	return
	// }

	// t.Log(cookie)


	cookie,err := rclient.Create()

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(cookie)



	cv := cookie.Value

	ret , err := rclient.IsExists(cv)

	if err != nil {
		t.Error(err)
		return
	}
	
	
	if *ret == DATA_NOT_FOUND {
		t.Error("key not exists")
		return
	}


	var kvval TestObject

	ret , err = rclient.Get(cv,suffix1,&kvval)

	if err != nil  || *ret == DATA_NOT_FOUND {
		t.Error("keyA not exists")
		return
	}

	t.Log(kvval)


	err = rclient.Set(cv,suffix2,TestObject{"testname2","testvalue2"})

	if err != nil {
		t.Error(err)
		return
	}


	cookie,err = rclient.Regenerate(cv)

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(cookie)


	cv = cookie.Value


	err = rclient.RemoveChild(cv,suffix1)

	if err != nil {
		t.Error(err)
		return
	}


	ret , err = rclient.Get(cv,suffix1,&kvval)

	if err!=nil {
		t.Error(err)
		return
	}

	if *ret == DATA_NOT_FOUND {
		t.Log("suffix1 not exists")
	}






	ret , err = rclient.Get(cv,suffix2,&kvval)

	if err != nil  || *ret == DATA_NOT_FOUND {
		t.Error("suffix2 not exists")
		return
	}


	t.Log(kvval)


	cookie , err = rclient.Delete(cv)

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(cookie)


}
