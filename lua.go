package main

import "C"
import (
    "glua/core"
    "os"
    "os/signal"
    "syscall"
    "time"
    "unsafe"
)

/*
#cgo CFLAGS: -I ${SRCDIR}/lua
#cgo windows,!llua LDFLAGS: -L${SRCDIR}/libs/windows -llua -lm -lws2_32
#cgo linux,!llua LDFLAGS: -L/usr/lib/x86_64-linux-gnu -llua5.4 -lm
#cgo linux CFLAGS: -I /usr/include/lua5.4

//#cgo darwin,!llua LDFLAGS: -L${SRCDIR}/libs/macos -llua -lm -fPIC
#cgo darwin,!llua LDFLAGS: -llua -lm -fPIC

#include <lua.h>
#include <lauxlib.h>
#include <stdlib.h>
extern int l_init(lua_State* L);
extern int l_poll(lua_State* L);
extern int l_disp(lua_State* L);

static int c_luaopen_glua(lua_State* L) {
    static const struct luaL_Reg nativeFuncLib [] = {
         {"init", l_init},
         {"poll", l_poll},
         {"disp", l_disp},
         {NULL, NULL}
    };
    luaL_newlib(L, nativeFuncLib);
    return 1;
}
#define MT_GOFUNCTION "Lua.GoFunction"
extern int golua_callgofunction(lua_State *L, int ud);
static void *testudata(lua_State *L, int ud, const char *tname) {
    void *p = lua_touserdata(L, ud);
    if (p != NULL) {
        if (lua_getmetatable(L, ud)) {
            luaL_getmetatable(L, tname);
            if (!lua_rawequal(L, -1, -2)) {
                p = NULL;
            }
            lua_pop(L, 2);
            return p;
        }
    }
    return NULL;
}
static int callback_function(lua_State* L) {
    unsigned int *fid = testudata(L, 1, MT_GOFUNCTION);
    if (fid == NULL) return 0;

    lua_remove(L,1);
    return golua_callgofunction(L, *fid);
}
static void clua_initstate(lua_State* L) {
    luaL_newmetatable(L, MT_GOFUNCTION);

    lua_pushliteral(L,"__call");
    lua_pushcfunction(L,&callback_function);
    lua_settable(L,-3);
}
static void clua_pushgofunction(lua_State* L, unsigned int fid) {
    unsigned int* fidptr = (unsigned int *)lua_newuserdata(L, sizeof(unsigned int));
    *fidptr = fid;
    luaL_getmetatable(L, MT_GOFUNCTION);
    lua_setmetatable(L, -2);
}
*/
import "C"

func main() {}

var (
    ch = make(chan core.Mail, 1000)
)

//export luaopen_glua
func luaopen_glua(L *C.lua_State) int {
    C.clua_initstate(L)
    go func() {
        sigs := make(chan os.Signal, 1)
        signal.Notify(sigs,
            syscall.SIGINT,
            syscall.SIGILL,
            syscall.SIGFPE,
            syscall.SIGSEGV,
            syscall.SIGTERM,
            syscall.SIGABRT)
        <-sigs
        close(ch)
        for {
            if len(ch) == 0 {
                break
            }
            time.Sleep(time.Millisecond)
        }
        os.Exit(0)
    }()

    return int(C.c_luaopen_glua(L))
}

type LuaGoFunction func(L *C.lua_State) C.int

var (
    registry = core.NewRegistry()
)

func pushGoFunction(L *C.lua_State, name string, fn LuaGoFunction) {
    fid := registry.Put(fn)
    C.clua_pushgofunction(L, C.uint(fid))

    Cname := C.CString(name)
    defer C.free(unsafe.Pointer(Cname))
    C.lua_setglobal(L, Cname)
}

//export l_init
func l_init(L *C.lua_State) C.int {
    return 0
}

//export golua_callgofunction
func golua_callgofunction(L *C.lua_State, id C.int) C.int {
    p := registry.Get(uint32(id))
    if fn, ok := p.(LuaGoFunction); ok {
        return fn(L)
    }
    return 0
}

//export l_poll
func l_poll(L *C.lua_State) C.int {
    m := <-ch
    core.Filter(ch, m.Id, m.Args...)
    return returnArgs(L, m.Get()...)
}

//export l_disp
func l_disp(L *C.lua_State) C.int {
    if C.lua_isnumber(L, 1) != 1 {
        return 0
    }
    args := checkArgs(L)
    id := int(C.lua_tointegerx(L, 1, nil))
    if len(args) > 0 {
        args = args[1:]
    }
    ch <- core.Mail{
        Id:   id,
        Args: args,
    }
    return 0
}

// #define LUA_TNIL		0
// #define LUA_TBOOLEAN		1
// #define LUA_TLIGHTUSERDATA	2
// #define LUA_TNUMBER		3
// #define LUA_TSTRING		4
// #define LUA_TTABLE		5
// #define LUA_TFUNCTION		6
// #define LUA_TUSERDATA		7
// #define LUA_TTHREAD		8
func checkArgs(L *C.lua_State) (args []interface{}) {
    n := int(C.lua_gettop(L))
    for i := 1; i <= n; i++ {
        t := C.lua_type(L, C.int(i))
        switch int(t) {
        case 0: // LUA_TNIL
            args = append(args, nil)
        case 1: // LUA_TBOOLEAN
            args = append(args, int(C.lua_toboolean(L, C.int(i))) == 1)
        case 2: // LUA_TLIGHTUSERDATA
            args = append(args, nil)
        case 3: // LUA_TNUMBER
            args = append(args, float64(C.lua_tonumberx(L, C.int(i), nil)))
        case 4: // LUA_TSTRING
            var sz C.size_t
            c_strPtr := C.luaL_checklstring(L, C.int(i), &sz)
            args = append(args, C.GoStringN(c_strPtr, C.int(sz)))
        case 5: // LUA_TTABLE
            args = append(args, nil)
        case 6: // LUA_TFUNCTION
            args = append(args, nil)
        case 7: // LUA_TUSERDATA
            args = append(args, nil)
        case 8: // LUA_TTHREAD
            args = append(args, nil)
        default:
            args = append(args, nil)
        }
    }
    return
}
func returnArgs(L *C.lua_State, args ...interface{}) C.int {
    for _, arg := range args {
        switch v := arg.(type) {
        case []byte:
            sz := C.size_t(len(v))
            data := C.CString(string(v))
            C.lua_pushlstring(L, data, sz)
            C.free(unsafe.Pointer(data))
        case string:
            sz := C.size_t(len(v))
            cs := C.CString(v)
            C.lua_pushlstring(L, cs, sz)
            C.free(unsafe.Pointer(cs))
        case int:
            C.lua_pushinteger(L, C.longlong(v))
        case int8:
            C.lua_pushinteger(L, C.longlong(v))
        case int16:
            C.lua_pushinteger(L, C.longlong(v))
        case int32:
            C.lua_pushinteger(L, C.longlong(v))
        case int64:
            C.lua_pushinteger(L, C.longlong(v))
        case uint8:
            C.lua_pushinteger(L, C.longlong(v))
        case uint16:
            C.lua_pushinteger(L, C.longlong(v))
        case uint32:
            C.lua_pushinteger(L, C.longlong(v))
        case uint64:
            C.lua_pushinteger(L, C.longlong(v))
        case uintptr:
            C.lua_pushinteger(L, C.longlong(v))
        case float32:
            C.lua_pushnumber(L, C.double(v))
        case float64:
            C.lua_pushnumber(L, C.double(v))
        default:
            C.lua_pushnil(L)
        }
    }
    return C.int(len(args))
}
