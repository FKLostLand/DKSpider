package FKUIWeb

import (
	"github.com/astaxie/beego/session"
)

var (
	G_SessionMgr = func() *session.Manager {
		sm, _ := session.NewManager("memory",
			`{"cookieName":"FKSession", "enableSetCookie,omitempty": true, 
					"secure": false, "sessionIDHashFunc": "sha1", "sessionIDHashKey": "", 
					"cookieLifeTime": 157680000, "providerConfig": ""}`)
		return sm
	}()
)
