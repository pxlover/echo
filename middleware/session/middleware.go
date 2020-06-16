/*

   Copyright 2016 Wenhui Shen <www.webx.top>

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

*/
package session

import (
	"errors"

	"github.com/admpub/sessions"
	"github.com/webx-top/echo"
)

const CookieMaxAgeKey = `CookieMaxAge`

func RememberMaxAge(c echo.Context, maxAge int) {
	if maxAge > 0 {
		c.CookieOptions().MaxAge = maxAge
		c.Session().Set(CookieMaxAgeKey, maxAge)
	} else {
                if c.Session().Get(CookieMaxAgeKey) != nil {
                       c.Session().Delete(CookieMaxAgeKey)
                }
        }
}

func Sessions(options *echo.SessionOptions, store sessions.Store) echo.MiddlewareFuncd {
	var newSession func(ctx echo.Context) echo.Sessioner
	if options == nil {
		newSession = func(ctx echo.Context) echo.Sessioner {
			return NewMySession(store, ctx.SessionOptions().Name, ctx)
		}
	} else {
		newSession = func(ctx echo.Context) echo.Sessioner {
			sessionOptions := options.Clone()
			ctx.SetSessionOptions(sessionOptions)
			return NewMySession(store, sessionOptions.Name, ctx)
		}
	}
	return func(h echo.Handler) echo.HandlerFunc {
		return func(c echo.Context) error {
			s := newSession(c)
			c.SetSessioner(s)
			s.SetPreSaveHook(func(c echo.Context) error{
				maxAge, ok := s.Get(CookieMaxAgeKey).(int)
				if !ok {
					return nil
				}
				if maxAge > 3600 && c.CookieOptions().MaxAge != maxAge {
					c.CookieOptions().MaxAge = maxAge
				}
				return nil
			})
			c.AddPreResponseHook(s.Save)
			err := h.Handle(c)
			if e := s.Save(); e != nil {
				if err != nil {
					err = errors.New("Multiple errors:\n1. " + err.Error() + "\n2. " + e.Error())
				} else {
					err = e
				}
				c.Logger().Error(err)
			}
			return err
		}
	}
}

func Middleware(options *echo.SessionOptions) echo.MiddlewareFuncd {
	store := StoreEngine(options)
	return Sessions(options, store)
}
