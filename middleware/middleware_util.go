// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"encoding/base64"
	"net"
	"net/http"
	"strings"
	"time"

	"git.hoogi.eu/snafu/go-blog/logger"
)

var locals = [...]net.IPNet{
	{
		IP:   net.IPv4(10, 0, 0, 0),
		Mask: net.CIDRMask(8, 32),
	},
	{
		IP:   net.IPv4(172, 16, 0, 0),
		Mask: net.CIDRMask(12, 32),
	},
	{
		IP:   net.IPv4(192, 168, 0, 0),
		Mask: net.CIDRMask(16, 32),
	},
}

func getIP(r *http.Request) string {
	xfo := r.Header.Get("X-Forwarded-For")
	xre := r.Header.Get("X-Real-IP")

	if len(xre) > 0 {
		ips := strings.Split(xre, ", ")

		for _, ip := range ips {
			parsedIP := net.ParseIP(ip)
			for _, local := range locals {
				if !local.Contains(parsedIP) && !parsedIP.IsLoopback() {
					return ip
				}
			}
		}
	}

	if len(xfo) > 0 {
		ips := strings.Split(xfo, ", ")

		for _, ip := range ips {
			parsedIP := net.ParseIP(ip)
			for _, local := range locals {
				if !local.Contains(parsedIP) && !parsedIP.IsLoopback() {
					return ip
				}
			}
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		logger.Log.Errorf("error while determining ip address from request %v", err)
	}

	return ip
}

func setCookie(rw http.ResponseWriter, name, path, data string) {
	c := &http.Cookie{
		Name:  name,
		Path:  path,
		Value: base64.StdEncoding.EncodeToString([]byte(data)),
	}

	http.SetCookie(rw, c)
}

func getFlash(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	c, err := r.Cookie(name)

	if err != nil {
		switch err {
		case http.ErrNoCookie:
			return "", nil
		default:
			return "", err
		}
	}

	value, err := base64.StdEncoding.DecodeString(c.Value)

	if err != nil {
		return "", err
	}

	//Remove temporary cookie immediately
	dc := &http.Cookie{
		Name:    name,
		MaxAge:  -1,
		Expires: time.Unix(1, 0),
		Path:    "/",
	}

	http.SetCookie(w, dc)

	return string(value), nil
}
