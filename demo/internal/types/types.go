// Code generated by goctl. DO NOT EDIT.
package types

type Request struct {
	Name string `path:"name,options=cancel|cron|once"`
}

type Response struct {
	Message string `json:"message"`
}
