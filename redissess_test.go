package redissess

import (
	"testing"
	"errors"
	// "fmt"
	"net/http"
)

const suffix1 = "suffix1"
const suffix2 = "suffix2"


var rclient *RedisClient


func createClient(t *testing.T) error {

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

	rclient = New(coninfo,sess)


	if rclient==nil {
		return errors.New("not connected")
	}



	return nil
}

type TestObject struct {
	Name string
	Value string
}










func TestExists(t *testing.T){

	err := createClient(t)

	defer func(){
		t.Log("close")
		rclient.Close()
	}()


	if err != nil {
		t.Fatal(err)
	}

	cookie,err := rclient.Create()

	if err != nil {
		t.Fatal(err)
	}

	t.Log(cookie)

	cv := cookie.Value

	ret , err := rclient.IsExists(cv)

	if err != nil {
		t.Fatal(err)
	}

	if *ret == DATA_NOT_FOUND {
		t.Fatal("key does not exists")
	}

	ret , err = rclient.IsExists(cv+"aaaa")

	if err != nil {
		t.Fatal(err)
	}

	if *ret == DATA_NOT_FOUND {
		t.Log("key does not exists")
	}


	cookie , err = rclient.Delete(cv)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(cookie)

	t.Log("END")
}


func TestRegenerate(t *testing.T){

	err := createClient(t)

	defer func(){
		t.Log("close")
		rclient.Close()
	}()


	if err != nil {
		t.Fatal(err)
	}


	test1 := TestObject{"testname","testvalue"}

	cookie,err := rclient.CreateAndSet(suffix1,test1)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(cookie)

	cv1 := cookie.Value

	cookie,err = rclient.Regenerate(cv1)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(cookie)

	cv2 := cookie.Value

	ret , err := rclient.IsExists(cv1)

	if err != nil {
		t.Fatal(err)
	}

	if *ret == DATA_NOT_FOUND {
		t.Log("key does not exists OK")
	}


	var test2 TestObject

	ret , err = rclient.Get(cv2,suffix1,&test2)

	if err != nil {
		t.Fatal(err)
	}

	if *ret == DATA_NOT_FOUND {
		t.Fatal("key does not exists")
	}

	cookie , err = rclient.Delete(cv2)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(cookie)
}


func TestSetAndGet(t *testing.T){

	err := createClient(t)

	defer func(){
		t.Log("close")
		rclient.Close()
	}()


	if err != nil {
		t.Fatal(err)
	}



	for i:=0 ;i<2 ; i++ {

		var cookie *http.Cookie
		var err error

		if i==0 {
			cookie,err = rclient.Create()
		} else {
			cookie,err = rclient.CreateAndSet(suffix1,TestObject{"testname","testvalue"})	
		}

		if err != nil {
			t.Fatal(err)
		}
	

		t.Logf("i=%d",i)
		t.Log(cookie)
	
		cv := cookie.Value
	
		err = rclient.Set(cv,suffix1,TestObject{"testname","testvalue"})
	
		if err != nil {
			t.Fatal(err)
		}	
	
		err = rclient.Set(cv,suffix2,TestObject{"testname2","testvalue2"})
	
		if err != nil {
			t.Fatal(err)
		}	
	
	
		err = rclient.RemoveChild(cv,suffix1)
	
		if err != nil {
			t.Fatal(err)
		}
	
	//get1
	
		var testobj TestObject
	
	
		ret , err := rclient.Get(cv,suffix1,&testobj)
	
		if err != nil {
			t.Fatal(err)
		}
	
		if *ret == DATA_NOT_FOUND {
			t.Log("suffix1 not exists")
		}
	
	//get2
	
		ret , err = rclient.Get(cv,suffix2,&testobj)
	
		if err != nil {
			t.Fatal(err)
		}
	
		if *ret == DATA_NOT_FOUND {
			t.Fatal("suffix2 not exists")
		}
	
	
		cookie , err = rclient.Delete(cv)
	
		if err != nil {
			t.Fatal(err)
		}
	

	}

}
