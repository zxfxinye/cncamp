package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang/glog"
)

func main() {
	flag.Set("v", "2")
	glog.V(2).Info("Starting http server...")
	http.HandleFunc("/", rootHandler)
	srv := http.Server{
		Addr:    ":80",
		Handler: http.DefaultServeMux,
	}
	//go func() {
	//	src := http.Server{
	//		Addr: ":8088",
	//		Handler: http.DefaultServeMux,
	//	}
	//	err:= src.ListenAndServe()
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//}()
	//go func() {
	//	src := http.Server{
	//		Addr: ":8088",
	//		Handler: http.DefaultServeMux,
	//	}
	//	err:= src.ListenAndServe()
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//}()
	sigs := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		//收到sigterm信号 10超时关闭服务
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			fmt.Println("暴力终止: ", err)
		}
		done <- struct{}{}
	}()
	err := srv.ListenAndServe()
	<-done
	if err != nil {
		log.Fatal(err)
	}

}
func rootHandler(w http.ResponseWriter, r *http.Request) {
	clientIp, err := GetIP(r)
	if err != nil {
		fmt.Println(err)
	}
	r.ParseForm()
	fmt.Fprintln(w, r.Form)
	for k,v := range r.PostForm{
		fmt.Println(k,": ",v)
	}
	//fmt.Printf("POST json: username=%s, password=%s\n", params["username"], params["password"])
	rspHeader := w.Header()
	rspHeader.Add("VERSION", os.Getenv("VERSION"))
	for k, v := range r.Header {
		rspHeader.Set(k, fmt.Sprintf("%s", v))
	}

	if r.URL.Path == "/healthz" {
		rspHeader.Add("StatusCode", fmt.Sprintf("%d", http.StatusOK))
		io.WriteString(w, fmt.Sprintf("%d\n", http.StatusOK))
	} else {
		user := r.URL.Query().Get("user")
		if user != "" {
			io.WriteString(w, fmt.Sprintf("hello [%s]\n", user))
		} else {
			io.WriteString(w, "hello [stranger]\n")
		}
	}
	glog.V(2).Infoln(clientIp, http.StatusOK)
}

func GetIP(r *http.Request) (string, error) {
	ip := r.Header.Get("X-Real-IP")
	if net.ParseIP(ip) != nil {
		return ip, nil
	}
	ip = r.Header.Get("X-Forward-For")
	for _, i := range strings.Split(ip, ",") {
		if net.ParseIP(i) != nil {
			return i, nil
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	if net.ParseIP(ip) != nil {
		return ip, nil
	}
	return "", errors.New("no valid ip found")
}
