package controllers

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/vesoft-inc/nebula-http-gateway/ccore/nebula/gateway/dao"
	"github.com/vesoft-inc/nebula-http-gateway/ccore/nebula/gateway/pool"
	"github.com/vesoft-inc/nebula-http-gateway/ccore/nebula/types"
	"github.com/vesoft-inc/nebula-http-gateway/common"
)

type DatabaseController struct {
	beego.Controller
}

type Response struct {
	Code    int       `json:"code"`
	Data    types.Any `json:"data"`
	Message string    `json:"message"`
}

type Request struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Address  string `json:"address"`
	Port     int    `json:"port"`

	/*
		if the request version field is "",
		will use `types.VersionHelper()` to infer a version
	*/
	Version string `json:"version"`
}

type ExecuteRequest struct {
	Gql       string              `json:"gql"`
	ParamList types.ParameterList `json:"paramList"`
}

type Data map[string]interface{}

func (this *DatabaseController) Connect() {
	var (
		res    Response
		params Request
	)
	json.Unmarshal(this.Ctx.Input.RequestBody, &params)

	//获取环境变量 DB_ADDRESS
	if params.Address == "" {
		params.Address = os.Getenv("DB_ADDRESS")
	}
	//获取环境变量 DB_PORT
	if params.Port == 0 {
		strPort := os.Getenv("DB_PORT")
		intPort, _ := strconv.Atoi(strPort)
		params.Port = intPort
	}

	info, err := dao.Connect(params.Address, params.Port, params.Username, params.Password)
	if err == nil {
		nsid := info.ClientID
		res.Code = 0
		res.Data = nsid
		this.Ctx.SetCookie("Secure", "true")
		this.Ctx.SetCookie("SameSite", "Strict")
		this.SetSession(beego.AppConfig.String("sessionkey"), nsid)
		//force session id to res.data
		res.Data = this.CruSession.SessionID()
		res.Message = "Login successfully"
	} else {
		res.Code = -1
		res.Message = err.Error()
	}

	this.Data["json"] = &res
	this.ServeJSON()
}

func (this *DatabaseController) Home() {
	var res Response
	res.Code = 0
	res.Data = "Run Successfully!"
	res.Message = "Welcome to nebula http gateway!"
	this.Data["json"] = &res
	this.ServeJSON()
}

func (this *DatabaseController) Disconnect() {
	var res Response
	nsid := this.GetSession(beego.AppConfig.String("sessionkey"))
	if nsid == nil {
		res.Code = -1
		res.Message = "No connection existed"
	} else {
		err := dao.Disconnect(nsid.(string))
		if err != nil {
			res.Code = -1
			res.Message = "Disconnect failed"
		} else {
			res.Code = 0
			res.Message = "Disconnect successfully"
		}
	}

	this.Data["json"] = &res
	this.ServeJSON()
}

func (this *DatabaseController) Execute() {
	var res Response
	var params ExecuteRequest
	nsid := this.GetSession(beego.AppConfig.String("sessionkey"))
	if nsid == nil {
		res.Code = -1
		res.Message = "connection refused for lack of session"
	} else {
		json.Unmarshal(this.Ctx.Input.RequestBody, &params)
		result, msg, err := dao.Execute(nsid.(string), params.Gql, params.ParamList)
		if msg != nil {
			if err == pool.SessionLostError {
				common.LogPanic(msg)
			} else {
				logs.Error(msg)
			}
		}

		if err == nil {
			res.Code = 0
			res.Data = &result
		} else {
			res.Code = -1
			res.Message = err.Error()
		}
	}
	this.Data["json"] = &res
	this.ServeJSON()
}
