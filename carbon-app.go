package main

//go:generate go-bindata -nomemcopy -pkg glue -o glue/generated_glue.go -prefix "./app" ./app

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"./glue"

	"github.com/DeedleFake/Go-PhysicsFS/physfs"
	"github.com/carbonsrv/carbon/modules/luaconf"
	"github.com/carbonsrv/carbon/modules/middleware"
	"github.com/carbonsrv/carbon/modules/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/pmylund/go-cache"
	"github.com/vifino/golua/lua"
	"github.com/vifino/luar"
	"golang.org/x/net/http2"
)

// General
var jobs *int

// Cache
var cfe *cache.Cache
var kvstore *cache.Cache

func cacheRead(c *cache.Cache, file string) (string, error) {
	res := ""
	data_tmp, found := c.Get(file)
	if found == false {
		data, err := fileRead(file)
		if err != nil {
			return "", err
		}
		c.Set(file, data, cache.DefaultExpiration)
	} else {
		debug("Using cache for %s" + file)
		res = data_tmp.(string)
	}
	return res, nil
}

// File system functions
var filesystem http.FileSystem

func initPhysFS(path string) http.FileSystem {
	err := physfs.Init()
	if err != nil {
		panic(err)
	}
	err = physfs.Mount(path, "/", true)
	if err != nil {
		panic(err)
	}
	return physfs.FileSystem()
}

/*func fileExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}*/
func fileRead(file string) (string, error) {
	if physfs.Exists(file) {
		f, err := filesystem.Open(file)
		defer f.Close()
		if err != nil {
			return "", err
		}
		fi, err := f.Stat()
		if err != nil {
			return "", err
		}
		r := bufio.NewReader(f)
		buf := make([]byte, fi.Size())
		_, err = r.Read(buf)
		if err != nil {
			return "", err
		}
		return string(buf), err
	}
	return "", errors.New("no such file or directory.")
}
func fileExists(file string) bool {
	return physfs.Exists(file)
}
func cacheFileExists(file string) bool {
	data_tmp, found := cfe.Get(file)
	if found == false {
		exists := fileExists(file)
		cfe.Set(file, exists, cache.DefaultExpiration)
		return exists
	} else {
		return data_tmp.(bool)
	}
}

// Logging
var doLog bool = false

func debug(str string) {
	if doLog {
		log.Print(str)
	}
}

// Server
func new_server() *gin.Engine {
	return gin.New()
}

func serve(srv http.Handler, en_http bool, en_https bool, en_http2 bool, bind string, binds string, cert string, key string) {
	end := make(chan bool)
	if en_http {
		go serveHTTP(srv, bind, en_http2)
	}
	if en_https {
		cert, _ := filepath.Abs(cert)
		key, _ := filepath.Abs(key)
		go serveHTTPS(srv, binds, en_http2, cert, key)
	}
	<-end
}
func serveHTTP(srv http.Handler, bind string, en_http2 bool) {
	s := &http.Server{
		Addr:           bind,
		Handler:        srv,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if en_http2 {
		http2.ConfigureServer(s, nil)
	}
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
func serveHTTPS(srv http.Handler, bind string, en_http2 bool, cert string, key string) {
	s := &http.Server{
		Addr:           bind,
		Handler:        srv,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if en_http2 {
		http2.ConfigureServer(s, nil)
	}
	err := s.ListenAndServeTLS(cert, key)
	if err != nil {
		panic(err)
	}
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

// Helper for environment
func int_from_env(key string, def int) int {
	var num = def
	tmp := os.Getenv(key)
	if tmp != "" {
		num, err := strconv.Atoi(tmp)
		assert(err)
		return num // to make go happy
	}
	return num
}
func bool_from_env(key string, def bool) bool {
	var boolean = def
	tmp := os.Getenv(key)
	if tmp != "" {
		boolean, err := strconv.ParseBool(tmp)
		assert(err)
		return boolean // to make go happy
	}
	return boolean
}
func string_from_env(key string, def string) string {
	var str = def
	tmp := os.Getenv(key)
	if tmp != "" {
		str = tmp
	}
	return str
}

// bindhook is the hook that registers our stuff
func bindhook(L *lua.State) {
	luar.Register(L, "carbon", luar.Map{
		"appdata": glue.GetGlue,
	})

	L.DoString(`
		local cache_key_prefix = "carbon:lua_module:"
		local cache_key_app = cache_key_prefix .. "app:"
		local cache_key_app_location = cache_key_prefix .. "app_location:"

		-- Load from compiled in app
		local function loadapp(name)
			local modname = tostring(name):gsub("%.", "/")
			local location = modname .. ".lua"
			local src = carbon.appdata(location)
			if src ~= "" then
				-- Compile and return the module
				local f, err = loadstring(src, location)
				if err then error(err, 0) end
				kvstore._set(cache_key_app..modname, string.dump(f))
				kvstore._set(cache_key_app_location..modname, location)
				return f
			end

			local location_init = modname .. "/init.lua"
			local src = carbon.appdata(location_init)
			if src ~= "" then
				-- Compile and return the module
				local f, err = loadstring(src, location)
				if err then error(err, 0) end
				kvstore._set(cache_key_app..modname, string.dump(f))
				kvstore._set(cache_key_app_location..modname, location_init)
				return f
			end
			return "\n\tno app asset '/" .. location .. "' (not compiled in)"..
				"\n\tno app asset '/" .. location_init .. "' (not compiled in)"
		end

		table.insert(package.loaders, 3, loadapp)
	`)
}

func main() {
	cfe = cache.New(5*time.Minute, 30*time.Second) // File-Exists Cache
	kvstore = cache.New(-1, -1)                    // Key-Value Storage

	var host = os.Getenv("CARBON_HOST")
	var port = int_from_env("CARBON_PORT", 80)
	var ports = int_from_env("CARBON_PORTS", 443)
	var cert = os.Getenv("CARBON_CERT")
	var key = os.Getenv("CARBON_KEY")
	var en_http = bool_from_env("CARBON_ENABLE_HTTP", true)
	var en_https = bool_from_env("CARBON_ENABLE_HTTPS", false)
	var en_http2 = bool_from_env("CARBON_ENABLE_HTTP2", false)

	wrkrs := 2
	if runtime.NumCPU() > 2 {
		wrkrs = runtime.NumCPU()
	}
	jobs := int_from_env("CARBON_STATES", wrkrs)
	var workers = int_from_env("CARBON_WORKERS", wrkrs)
	var webroot = string_from_env("CARBON_ROOT", ".")

	// Do debug!
	doDebug := bool_from_env("CARBON_DEBUG", false)
	// Middleware options
	useRecovery := bool_from_env("CARBON_RECOVERY", false)
	useLogger := bool_from_env("CARBON_LOGGER", true)

	args := os.Args

	if en_https {
		if key == "" || cert == "" {
			panic("Need to have a Key and a Cert defined.")
		}
	}

	runtime.GOMAXPROCS(workers)

	root, _ := filepath.Abs(webroot)
	physroot_path := root
	filesystem = initPhysFS(physroot_path)

	defer physfs.Deinit()
	go scheduler.Run()                        // Run the scheduler.
	go middleware.Preloader()                 // Run the Preloader.
	middleware.Init(jobs, cfe, kvstore, root) // Run init sequence.

	if doDebug == false {
		gin.SetMode(gin.ReleaseMode)
	}

	if useLogger {
		doLog = true
	}

	if flag.Arg(1) == "" {
		args = make([]string, 0)
	} else {
		args = args[1:]
	}

	script_src, err := glue.GetGlue("app.lua")
	if err != nil {
		script_src, err = glue.GetGlue("init.lua")
		if err != nil {
			fmt.Println("Error: Compiled Bundle does not contain 'app.lua' or 'init.lua'. No idea what to run. Aborting.")
			os.Exit(1)
		}
	}
	L := luaconf.Setup(args, cfe, webroot, useRecovery, useLogger, false, func(srv *gin.Engine) {
		serve(srv, en_http, en_https, en_http2, host+":"+strconv.Itoa(port), host+":"+strconv.Itoa(ports), cert, key)
	}, bindhook)

	err = L.DoString(script_src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	L.Close()
}
