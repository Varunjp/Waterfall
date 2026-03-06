package redisclient

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
)

type OTPRepo struct {
	redis *redis.Client
}

func NewOTPRepo(rdb *redis.Client) *OTPRepo {
	return &OTPRepo{
		redis: rdb,
	}
}

func (r *OTPRepo) StoreOTP(email,otp string) error {

	key := "otp:reset:"+email 

	data := map[string]interface{} {
		"otp": otp,
		"attempts":0,
	}

	jsonData,_ := json.Marshal(data)

	return r.redis.Set(key,jsonData,5*time.Minute).Err()
}

func (r *OTPRepo) GetOTP(email string)(string,int,error) {

	key := "otp:reset:"+email 

	val,err := r.redis.Get(key).Result()
	if err != nil {
		return "",0,err 
	}

	var data struct {
		Otp  string `json:"otp"`
		Attemps int `json:"attempts"`
	}

	json.Unmarshal([]byte(val),&data)
	return data.Otp,data.Attemps,nil 
}

func (r *OTPRepo) IncrementAttempt(email string) error {

	key := "otp:reset:"+email 

	val,err := r.redis.Get(key).Result()
	if err != nil {
		return err 
	}

	var data struct {
		Otp   string `json:"otp"`
		Attempts int `json:"attempts"`
	}

	json.Unmarshal([]byte(val),&data)

	data.Attempts++

	newVal,_ := json.Marshal(data)

	return r.redis.Set(key,newVal,5*time.Minute).Err()
}

func (r *OTPRepo) CanResend(email string) bool {

	key := "otp:reset:cooldown:" + email

	exists, _ := r.redis.Exists(key).Result()

	return exists == 0
}

func (r *OTPRepo) SetCooldown(email string) {

	key := "otp:reset:cooldown:" + email

	r.redis.Set(key, "1", 30*time.Second)
}