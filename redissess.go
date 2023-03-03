package redissess

import (
	"context"
	"github.com/go-redis/redis/v8"
	"encoding/json"
	"io"
	"strings"
	"crypto/rand"
	"encoding/base32"
	"time"
	"net/http"
)



const DATA_NOT_FOUND = 0
const DATA_FOUND = 1

type ConnInfo struct {
	Addr string
	DB int
}


type SessionParam struct{
	// Cookie http.Cookie

	Name string
	Path string
	Domain string
	MaxAge int
	Secure bool
	HttpOnly bool
	SameSite http.SameSite
	Lifetime *int
}

type RedisClient struct {
	Client *redis.Client
}

var sessparam SessionParam

func New(conn ConnInfo,param SessionParam) *RedisClient {

	var rclient RedisClient

	nc := redis.NewClient(&redis.Options{
				Addr: conn.Addr,
				DB:   conn.DB,
			})

	if param.Lifetime == nil {
		param.Lifetime = &param.MaxAge
	}

	sessparam = param

	ctx := context.Background()

	_,err := nc.Ping(ctx).Result()

	if err != nil {
		return nil
	}

	rclient.Client = nc

	return &rclient
}

func (rc *RedisClient) Close() {
	rc.Client.Close()
}


func createSessid() (string, error) {
	k := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(k), "="), nil
}

func  (rc *RedisClient) Create() (*http.Cookie,error) {

	str,err := createSessid()

	if err != nil {
		return nil,err
	}

	err = rc.Client.Set(context.Background(),
						str,
						nil,
						time.Second*time.Duration(*sessparam.Lifetime),
					).Err()

	if err != nil {
		return nil,err
	}



	return newCookie(str),nil
}




func (rc *RedisClient) CreateAndSet(suffix string,obj interface{}) (*http.Cookie,error) {

	str,err := createSessid()

	if err != nil {
		return nil,err
	}

	err = rc.Set(str,suffix,obj)

	if err != nil {
		return nil,err
	}

	return newCookie(str),nil
}



func (rc *RedisClient) Set(key string,suffix string,obj interface{}) error {

	jobj,err := json.Marshal(obj)

	 if err != nil {
	 	return err
	 }


	ctx := context.Background()

	// err = rc.Client.Watch(ctx,func(tx *redis.Tx) error {
		
	// 		// data, err := tx.Get(ctx, key).Result()
	// 		_, err := tx.Get(ctx, key).Result()

	// 		if err != nil && err != redis.Nil {
	// 			return err
	// 		}

	// 		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
	// 			pipe.HSet(ctx,key,suffix,jobj)

	// 			pipe.Expire(ctx,
	// 			 			key,
	// 			 			time.Second*time.Duration(*sessparam.Lifetime),
	// 			 		)

	// 			return nil
	// 		})

	// 		return err

	// 	},key,
	// )

//type

	ret ,err := rc.Client.Type(ctx,key).Result()

	if err != nil {
		return err
  	}

	if ret != "hash" {

		if err := rc.Client.Del(ctx,key).Err(); err!=nil{
			return err
		}

		err = rc.Client.HSetNX(ctx, 
						key,
						suffix,
						jobj,
				).Err()

		if err != nil {
			return err
		}
		

	}




	err = rc.Client.HSet(ctx, 
					   key,
					   suffix,
					   jobj,
			).Err()

	if err != nil {
  		return err
	}

	if err = rc.setExpire(ctx,key) ; err != nil {
		return err
	}

	return nil
}


func (rc *RedisClient) Get(key string,suffix string,obj interface{}) (*int,error) {
	
	ret := DATA_NOT_FOUND

	ctx := context.Background()

	data,err := rc.Client.HGet(ctx,key,suffix).Result()

	if err == redis.Nil {
		return &ret,nil
	} 
	
	if err != nil {
		return nil,err
	}

	if obj != nil {
		err = json.Unmarshal([]byte(data),obj)

		if err != nil {
			return nil,err
		}
	}

	if err = rc.setExpire(ctx,key) ; err != nil {

		return nil,err
	}


	ret = DATA_FOUND

	return &ret,nil
}




func (rc *RedisClient) setExpire(ctx context.Context, key string) error {

	be := rc.Client.Expire(ctx,
						   key,
						   time.Second*time.Duration(*sessparam.Lifetime),
	 					  )

	if be.Err() != nil {
		return be.Err()
	}

	return nil

}

func (rc *RedisClient) IsExists(key string) (*int,error) {

	ctx := context.Background()

	n,err := rc.Client.Exists(ctx,key).Result()

	if err!=nil {
		return nil,err
	} 

	ret := DATA_NOT_FOUND

	if n > 0 {
		ret = DATA_FOUND

		if err = rc.setExpire(ctx,key) ; err != nil {
			return nil,err
		}

	}

	return &ret,nil	 
}

func (rc *RedisClient) Delete(key string) (*http.Cookie,error) {

	ctx := context.Background()

	if err := rc.Client.Del(ctx,key).Err(); err!=nil{
		return nil,err
	}

	expiresAt := time.Now()
	expiresAt = expiresAt.Add(time.Second*(-10))

	return &http.Cookie{
		Name : sessparam.Name,
		Value : "",
		Domain: sessparam.Domain,
		Path : sessparam.Path,
		Expires:expiresAt,
		MaxAge : -1,
		HttpOnly : sessparam.HttpOnly,
		Secure : sessparam.Secure,
		SameSite: sessparam.SameSite,
	},nil

}

func newCookie(value string ) *http.Cookie {

	return &http.Cookie{
		Name : sessparam.Name,
		Value : value,
		Domain: sessparam.Domain,
		Path : sessparam.Path,
		MaxAge : sessparam.MaxAge,
		HttpOnly : sessparam.HttpOnly,
		Secure : sessparam.Secure,
		SameSite: sessparam.SameSite,
	}
}


 func (rc *RedisClient) Regenerate(currentkey string) (*http.Cookie,error) {

	dbnbr := rc.Client.Options().DB

	destkey,err := createSessid()

	if err != nil {
		return nil,err
	}	

	ctx := context.Background()

	n,err := rc.Client.Copy(ctx,currentkey,destkey,dbnbr,false).Result()

	if err != nil || n==0 {
		return nil,err
	}

	_,err = rc.Delete(currentkey)

	if err != nil {
		return nil,err
	}

	return newCookie(destkey),nil
 }


 func (rc *RedisClient) RemoveChild(key string,suffix string) error {
 
	err := rc.Client.HDel(context.Background(),key,suffix).Err()

	if err!=nil {
		return err
	}

	return nil
}
